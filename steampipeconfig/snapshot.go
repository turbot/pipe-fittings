package steampipeconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
// but each panel is converted and written separately, and the rows of a SnapshotRowsPanel one at a time.
// This keeps memory use proportional to the largest panel or row, not to the whole snapshot.
func (s *SteampipeSnapshot) WriteStrippedJson(w io.Writer, indent bool) error {
	fields, err := s.cloudSnapshotFields()
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(w)
	jw := &jsonWriter{w: bw, indent: indent}

	jw.write("{")
	for i, key := range slices.Sorted(maps.Keys(fields)) {
		jw.key(i, key, 1)
		if key == "panels" && s.Panels != nil {
			s.writeStrippedPanels(jw)
		} else {
			jw.value(fields[key], 1)
		}
	}
	jw.close(len(fields), 0, "}")

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
	jw.write("{")
	for i, name := range names {
		jw.key(i, name, 2)
		if rowsPanel, ok := s.Panels[name].(SnapshotRowsPanel); ok {
			jw.setErr(writeStrippedRowsPanel(jw, rowsPanel, 2))
			continue
		}
		panelJson, err := strippedPanelJson(s.Panels[name])
		jw.setErr(err)
		jw.value(panelJson, 2)
	}
	jw.close(len(names), 1, "}")
}

// writeStrippedRowsPanel writes the stripped panel at the given depth, converting its data rows one at a time
func writeStrippedRowsPanel(jw *jsonWriter, p SnapshotRowsPanel, depth int) error {
	headJson, err := strippedPanelJson(p.PanelWithoutRows())
	if err != nil {
		return err
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(headJson, &members); err != nil || members == nil {
		return fmt.Errorf("panel without rows must be a JSON object: %s", headJson)
	}
	var data map[string]json.RawMessage
	if dataJson, ok := members["data"]; ok {
		if err := json.Unmarshal(dataJson, &data); err != nil {
			return fmt.Errorf("panel data must be a JSON object or null: %w", err)
		}
	}
	// no data object, so no rows to write
	if data == nil {
		jw.value(headJson, depth)
		return nil
	}
	if _, ok := data["rows"]; ok {
		return fmt.Errorf("panel without rows has data rows")
	}

	keys := slices.Sorted(maps.Keys(members))
	jw.write("{")
	for i, key := range keys {
		jw.key(i, key, depth+1)
		if key == "data" {
			writeDataWithRows(jw, data, p, depth+1)
		} else {
			jw.value(members[key], depth+1)
		}
	}
	jw.close(len(keys), depth, "}")
	return nil
}

// writeDataWithRows writes the data object at the given depth, with the panel rows added as "rows"
func writeDataWithRows(jw *jsonWriter, data map[string]json.RawMessage, p SnapshotRowsPanel, depth int) {
	keys := slices.Sorted(slices.Values(append(slices.Collect(maps.Keys(data)), "rows")))
	jw.write("{")
	for i, key := range keys {
		jw.key(i, key, depth+1)
		if key != "rows" {
			jw.value(data[key], depth+1)
			continue
		}
		jw.write("[")
		count := 0
		jw.setErr(p.EachRow(func(row any) error {
			rowJson, err := genericJson(row)
			if err != nil {
				return err
			}
			jw.next(count, depth+2)
			jw.value(rowJson, depth+2)
			count++
			return jw.err
		}))
		jw.close(count, depth+1, "]")
	}
	jw.close(len(keys), depth, "}")
}

// toGeneric converts v to the generic form (maps, slices, float64 etc.) that a cloud snapshot holds
func toGeneric(v any) (any, error) {
	vJson, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var res any
	err = json.Unmarshal(vJson, &res)
	return res, err
}

// genericJson returns the JSON of v in the generic form that a cloud snapshot holds
func genericJson(v any) ([]byte, error) {
	res, err := toGeneric(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(res)
}

// strippedPanelJson converts the panel to the generic form used in a cloud snapshot, strips it and returns its JSON
func strippedPanelJson(p any) ([]byte, error) {
	panel, err := toGeneric(p)
	if err != nil {
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

// jsonWriter writes JSON one object member or array element at a time, in the same format as
// json.Marshal, or json.MarshalIndent with an empty prefix and a two space indent.
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

// setErr keeps err if it is the first error
func (j *jsonWriter) setErr(err error) {
	if j.err == nil {
		j.err = err
	}
}

// close writes the closing bracket of an object or array with the given member count, at the given depth
func (j *jsonWriter) close(members, depth int, bracket string) {
	if j.indent && members > 0 {
		j.write("\n" + strings.Repeat(jsonIndent, depth))
	}
	j.write(bracket)
}

// next writes the separator before member or element i, at the given depth
func (j *jsonWriter) next(i, depth int) {
	if i > 0 {
		j.write(",")
	}
	if j.indent {
		j.write("\n" + strings.Repeat(jsonIndent, depth))
	}
}

// key writes the separator before member i, then the member key, at the given depth
func (j *jsonWriter) key(i int, key string, depth int) {
	j.next(i, depth)
	keyJson, err := json.Marshal(key)
	if err != nil {
		j.setErr(err)
		return
	}
	if j.indent {
		j.write(string(keyJson) + ": ")
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
