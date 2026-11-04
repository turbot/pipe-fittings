package steampipeconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	steampipecloud "github.com/turbot/pipes-sdk-go"
)

var SteampipeSnapshotSchemaVersion int64 = 20240607

// propertiesToStrip are removed from each panel, both at the top level and under 'properties'
var propertiesToStrip = []string{
	"sql",
	"source_definition",
	"search_path",
	"search_path_prefix"}

type SteampipeSnapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	Panels        map[string]SnapshotPanel `json:"panels"`
	Inputs        map[string]interface{}   `json:"inputs"`
	Variables     map[string]string        `json:"variables"`
	SearchPath    []string                 `json:"search_path"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Layout        *SnapshotTreeNode        `json:"layout"`
	FileNameRoot  string                   `json:"-"`
	Title         string                   `json:"-"`
	Metadata      map[string]interface{}   `json:"metadata"`
}

// IsExportSourceData implements ExportSourceData
func (*SteampipeSnapshot) IsExportSourceData() {}

func (s *SteampipeSnapshot) AsCloudSnapshot() (*steampipecloud.WorkspaceSnapshotData, error) {
	jsonbytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	res := &steampipecloud.WorkspaceSnapshotData{}
	if err := json.Unmarshal(jsonbytes, res); err != nil {
		return nil, err
	}

	return res, nil
}

// AsStrippedJson returns the snapshot in cloud snapshot format, with propertiesToStrip removed from each panel.
func (s *SteampipeSnapshot) AsStrippedJson(indent bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.WriteStrippedJson(&buf, indent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteStrippedJson writes the snapshot to w in cloud snapshot format, with propertiesToStrip removed from each panel.
//
// The output is the same as marshalling the whole snapshot to a WorkspaceSnapshotData and back,
// but each panel is converted and written separately. This keeps memory use proportional to the
// largest panel, not to the whole snapshot.
func (s *SteampipeSnapshot) WriteStrippedJson(w io.Writer, indent bool) error {
	fields, err := s.cloudSnapshotFields()
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(w)
	jw := &jsonWriter{w: bw, indent: indent}

	jw.openObject()
	for i, key := range slices.Sorted(maps.Keys(fields)) {
		jw.key(i, key, 1)
		if key == "panels" && s.Panels != nil {
			s.writeStrippedPanels(jw)
		} else {
			jw.value(fields[key], 1)
		}
	}
	jw.closeObject(len(fields), 0)

	if jw.err != nil {
		return jw.err
	}
	return bw.Flush()
}

// cloudSnapshotFields converts every field except the panels to cloud snapshot format.
// It returns the JSON of each top level field, keyed by field name.
func (s *SteampipeSnapshot) cloudSnapshotFields() (map[string]json.RawMessage, error) {
	head := *s
	head.Panels = nil
	cloudHead, err := head.AsCloudSnapshot()
	if err != nil {
		return nil, err
	}
	headJson, err := json.Marshal(cloudHead)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(headJson, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func (s *SteampipeSnapshot) writeStrippedPanels(jw *jsonWriter) {
	names := slices.Sorted(maps.Keys(s.Panels))
	jw.openObject()
	for i, name := range names {
		panelJson, err := strippedPanelJson(s.Panels[name])
		if err != nil {
			jw.err = err
			return
		}
		jw.key(i, name, 2)
		jw.value(panelJson, 2)
	}
	jw.closeObject(len(names), 1)
}

// strippedPanelJson converts the panel to the generic form used in a cloud snapshot, strips it and returns its JSON
func strippedPanelJson(p SnapshotPanel) ([]byte, error) {
	panelJson, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var panel any
	if err := json.Unmarshal(panelJson, &panel); err != nil {
		return nil, err
	}
	stripPanel(panel)
	return json.Marshal(panel)
}

func StripSnapshot(snapshot *steampipecloud.WorkspaceSnapshotData) error {
	for _, p := range snapshot.Panels {
		stripPanel(p)
	}
	return nil
}

func stripPanel(p any) {
	panel, ok := p.(map[string]any)
	if !ok {
		return
	}
	properties, _ := panel["properties"].(map[string]any)
	for _, property := range propertiesToStrip {
		// look both at top level and under properties
		delete(panel, property)
		if properties != nil {
			delete(properties, property)
		}
	}
}

// jsonWriter writes a JSON object one member at a time, in the same format as json.Marshal,
// or json.MarshalIndent with an empty prefix and a two space indent.
// It keeps the first error and ignores later writes.
type jsonWriter struct {
	w      *bufio.Writer
	indent bool
	err    error
}

const jsonIndent = "  "

func (j *jsonWriter) write(s string) {
	if j.err == nil {
		_, j.err = j.w.WriteString(s)
	}
}

func (j *jsonWriter) openObject() {
	j.write("{")
}

// closeObject closes an object with the given member count, at the given depth
func (j *jsonWriter) closeObject(members, depth int) {
	if j.indent && members > 0 {
		j.write("\n" + strings.Repeat(jsonIndent, depth))
	}
	j.write("}")
}

// key writes the separator before member i, then the member key, at the given depth
func (j *jsonWriter) key(i int, key string, depth int) {
	if i > 0 {
		j.write(",")
	}
	keyJson, err := json.Marshal(key)
	if err != nil {
		j.err = err
		return
	}
	if j.indent {
		j.write("\n" + strings.Repeat(jsonIndent, depth) + string(keyJson) + ": ")
	} else {
		j.write(string(keyJson) + ":")
	}
}

// value writes compact JSON, indenting it for the given depth if required
func (j *jsonWriter) value(compact []byte, depth int) {
	if j.err != nil {
		return
	}
	if !j.indent {
		_, j.err = j.w.Write(compact)
		return
	}
	var buf bytes.Buffer
	if j.err = json.Indent(&buf, compact, strings.Repeat(jsonIndent, depth), jsonIndent); j.err == nil {
		_, j.err = buf.WriteTo(j.w)
	}
}
