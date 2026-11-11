package steampipeconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"strings"
	"testing"
	"time"
)

// legacyAsStrippedJson is the implementation of AsStrippedJson before it streamed panels.
// WriteStrippedJson must produce exactly the same bytes.
func legacyAsStrippedJson(s *SteampipeSnapshot, indent bool) ([]byte, error) {
	res, err := s.AsCloudSnapshot()
	if err != nil {
		return nil, err
	}
	if err = StripSnapshot(res); err != nil {
		return nil, err
	}
	if indent {
		return json.MarshalIndent(res, "", "  ")
	}
	return json.Marshal(res)
}

type testPanel struct {
	Name          string                   `json:"name"`
	Title         string                   `json:"title,omitempty"`
	PanelType     string                   `json:"panel_type"`
	SQL           string                   `json:"sql,omitempty"`
	Source        string                   `json:"source_definition,omitempty"`
	SearchPath    []string                 `json:"search_path,omitempty"`
	Properties    map[string]any           `json:"properties,omitempty"`
	Tags          map[string]string        `json:"tags,omitempty"`
	Data          *testData                `json:"data,omitempty"`
	Summary       map[string]int           `json:"summary,omitempty"`
	Duration      time.Duration            `json:"duration"`
	Error         error                    `json:"-"`
	Children      []map[string]interface{} `json:"children,omitempty"`
	Documentation *string                  `json:"documentation"`
}

func (*testPanel) IsSnapshotPanel() {}

type testData struct {
	Columns []map[string]string `json:"columns"`
	Rows    []map[string]any    `json:"rows"`
}

type mapPanel map[string]any

func (mapPanel) IsSnapshotPanel() {}

type failingPanel struct{}

func (failingPanel) IsSnapshotPanel() {}
func (failingPanel) MarshalJSON() ([]byte, error) {
	return nil, errors.New("panel cannot be marshalled")
}

// structRowsPanel is a testPanel whose data rows are written one at a time
type structRowsPanel struct {
	*testPanel
}

type testColumns struct {
	Columns []map[string]string `json:"columns"`
}

func (p structRowsPanel) PanelWithoutRows() any {
	var data *testColumns
	if p.Data != nil {
		data = &testColumns{Columns: p.Data.Columns}
	}
	return struct {
		*testPanel
		Data *testColumns `json:"data,omitempty"`
	}{p.testPanel, data}
}

func (p structRowsPanel) EachRow(f func(row any) error) error {
	for _, row := range p.Data.Rows {
		if err := f(row); err != nil {
			return err
		}
	}
	return nil
}

// mapRowsPanel is a panel head and its rows. It marshals in full: the head with the rows added to its data object.
type mapRowsPanel struct {
	head map[string]any
	rows []any
	// err is returned by EachRow after the rows
	err error
	// failFullMarshal makes MarshalJSON fail, to show that the whole panel is never marshalled
	failFullMarshal bool
}

func (*mapRowsPanel) IsSnapshotPanel() {}

func (p *mapRowsPanel) PanelWithoutRows() any { return p.head }

func (p *mapRowsPanel) EachRow(f func(row any) error) error {
	for _, row := range p.rows {
		if err := f(row); err != nil {
			return err
		}
	}
	return p.err
}

func (p *mapRowsPanel) MarshalJSON() ([]byte, error) {
	if p.failFullMarshal {
		return nil, errors.New("the whole panel must not be marshalled")
	}
	full := maps.Clone(p.head)
	if data, ok := full["data"].(map[string]any); ok {
		data = maps.Clone(data)
		data["rows"] = append([]any{}, p.rows...)
		full["data"] = data
	}
	return json.Marshal(full)
}

// rowsSnapshot is testSnapshot with SnapshotRowsPanel panels added
func rowsSnapshot() *SteampipeSnapshot {
	s := testSnapshot()
	columns := []map[string]string{{"name": "reason", "data_type": "TEXT"}, {"name": "region", "data_type": "TEXT"}}
	structRows := []map[string]any{
		{"reason": "bucket <b> & \"quoted\"", "resource": "arn:aws:s3:::b1", "status": "alarm", "region": "us-east-1"},
		{"reason": "unicode é ü 日本 \u2028", "resource": "", "status": "ok", "region": "eu-west-1"},
	}
	mapRows := []any{
		map[string]any{"reason": "<b>", "count": int64(1<<60 + 1), "ratio": 0.1, "big": 1e21, "small": 1e-7, "nil": nil},
		map[string]any{"nested": map[string]any{"z": 1, "a": []any{1, "two", nil, 3.5}}, "f": math.MaxFloat64},
		[]any{"a row that is not an object", 1},
		"a string row",
		nil,
	}
	maps.Copy(s.Panels, map[string]SnapshotPanel{
		"mod.control.rows_struct": structRowsPanel{&testPanel{
			Name: "mod.control.rows_struct", Title: "Rows <struct>", PanelType: "control",
			SQL: "select 1", Source: "control {}", SearchPath: []string{"aws"},
			Properties: map[string]any{"sql": "select 1", "search_path_prefix": "p", "severity": "high"},
			Data:       &testData{Columns: columns, Rows: structRows},
			Summary:    map[string]int{"alarm": 1, "ok": 1},
		}},
		"mod.control.rows_struct_empty":   structRowsPanel{&testPanel{Name: "mod.control.rows_struct_empty", PanelType: "control", Data: &testData{Columns: columns, Rows: []map[string]any{}}}},
		"mod.control.rows_struct_no_data": structRowsPanel{&testPanel{Name: "mod.control.rows_struct_no_data", PanelType: "control"}},
		"mod.control.rows_map": &mapRowsPanel{
			head: map[string]any{
				"name": "mod.control.rows_map", "sql": "select 2", "source_definition": "x",
				"properties": map[string]any{"source_definition": "y", "value": 42},
				"data":       map[string]any{"columns": columns, "a_before_rows": true, "z_after_rows": map[string]any{"k": "v"}},
			},
			rows: mapRows,
		},
		"mod.control.rows_map_null_data": &mapRowsPanel{head: map[string]any{"name": "mod.control.rows_map_null_data", "data": nil}},
		"mod.control.rows_map_no_data":   &mapRowsPanel{head: map[string]any{"name": "mod.control.rows_map_no_data"}},
		"mod.control.rows_map_empty":     &mapRowsPanel{head: map[string]any{"name": "mod.control.rows_map_empty", "data": map[string]any{}}},
	})
	return s
}

func testSnapshot() *SteampipeSnapshot {
	doc := "# Docs <with> &   line separator"
	rows := []map[string]any{
		{"reason": "bucket <b> & \"quoted\"", "resource": "arn:aws:s3:::b1", "status": "alarm", "region": "us-east-1"},
		{"reason": "ok", "resource": "arn:aws:s3:::b2", "status": "ok", "count": int64(1 << 60), "ratio": 0.1, "big": 1e21, "neg": math.Copysign(0, -1), "nil": nil},
		{"reason": "unicode é ü 日本  ", "resource": "", "status": "error", "nested": map[string]any{"z": 1, "a": []any{1, "two", nil, 3.5}}},
		{"reason": "max", "resource": "r", "status": "info", "f": math.MaxFloat64, "small": 1e-7, "i32": int32(-5)},
	}
	start := time.Date(2026, 11, 4, 17, 26, 5, 123456789, time.UTC)
	return &SteampipeSnapshot{
		SchemaVersion: fmt.Sprintf("%d", SteampipeSnapshotSchemaVersion),
		Panels: map[string]SnapshotPanel{
			"mod.control.c1": &testPanel{
				Name: "mod.control.c1", Title: "C1 <title>", PanelType: "control",
				SQL: "select 1", Source: "control \"c1\" {}", SearchPath: []string{"aws"},
				Properties:    map[string]any{"sql": "select 1", "source_definition": "x", "search_path": []string{"a"}, "search_path_prefix": "p", "severity": "high", "name": "c1"},
				Tags:          map[string]string{"service": "AWS/S3", "a": "b"},
				Data:          &testData{Columns: []map[string]string{{"name": "reason", "data_type": "TEXT"}}, Rows: rows},
				Summary:       map[string]int{"alarm": 1, "ok": 1, "error": 1, "info": 1, "skip": 0},
				Duration:      1500 * time.Millisecond,
				Documentation: &doc,
			},
			"mod.benchmark.b1":  &testPanel{Name: "mod.benchmark.b1", PanelType: "benchmark", Children: []map[string]interface{}{{"name": "mod.control.c1"}}},
			"mod.control.empty": &testPanel{Name: "mod.control.empty", PanelType: "control", Data: &testData{}, Properties: map[string]any{}},
			"mod.card.generic":  mapPanel{"name": "mod.card.generic", "sql": "select 2", "properties": map[string]any{"value": 42, "source_definition": "y"}},
			"mod.text.<html>&":  mapPanel{"name": "mod.text.<html>&", "value": "<p>hi</p>"},
		},
		Inputs:     map[string]interface{}{"input.region": "us-east-1", "input.n": 3},
		Variables:  map[string]string{"var.a": "1", "var.b": "<b>"},
		SearchPath: []string{"aws", "public"},
		StartTime:  start,
		EndTime:    start.Add(21 * time.Minute),
		Layout: &SnapshotTreeNode{
			Name:     "mod.benchmark.b1",
			NodeType: "benchmark",
			Children: []*SnapshotTreeNode{{Name: "mod.control.c1", NodeType: "control"}, {Name: "mod.control.empty", NodeType: "control"}},
		},
		Title:        "Snapshot <title>",
		FileNameRoot: "mod.benchmark.b1",
		Metadata:     map[string]interface{}{"view": map[string]any{"group_by": []any{"benchmark"}}},
	}
}

func TestWriteStrippedJson_MatchesLegacy(t *testing.T) {
	full := testSnapshot()

	noPanels := testSnapshot()
	noPanels.Panels = map[string]SnapshotPanel{}

	nilPanels := testSnapshot()
	nilPanels.Panels = nil

	minimal := &SteampipeSnapshot{SchemaVersion: "20240607"}

	cases := map[string]*SteampipeSnapshot{
		"full":        full,
		"no panels":   noPanels,
		"nil panels":  nilPanels,
		"minimal":     minimal,
		"rows panels": rowsSnapshot(),
	}
	for name, s := range cases {
		for _, indent := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/indent=%v", name, indent), func(t *testing.T) {
				want, err := legacyAsStrippedJson(s, indent)
				if err != nil {
					t.Fatalf("legacy: %v", err)
				}

				var buf bytes.Buffer
				if err := s.WriteStrippedJson(&buf, indent); err != nil {
					t.Fatalf("WriteStrippedJson: %v", err)
				}
				if got := buf.Bytes(); !bytes.Equal(got, want) {
					t.Fatalf("WriteStrippedJson output differs from legacy\n got: %s\nwant: %s", got, want)
				}

				got, err := s.AsStrippedJson(indent)
				if err != nil {
					t.Fatalf("AsStrippedJson: %v", err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("AsStrippedJson output differs from legacy\n got: %s\nwant: %s", got, want)
				}
			})
		}
	}
}

func TestWriteStrippedJson_StripsProperties(t *testing.T) {
	out, err := testSnapshot().AsStrippedJson(false)
	if err != nil {
		t.Fatal(err)
	}
	var res struct {
		Panels map[string]map[string]any `json:"panels"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatal(err)
	}
	for name, panel := range res.Panels {
		props, _ := panel["properties"].(map[string]any)
		for _, p := range propertiesToStrip {
			if _, ok := panel[p]; ok {
				t.Errorf("panel %s still has %s", name, p)
			}
			if _, ok := props[p]; ok {
				t.Errorf("panel %s still has properties.%s", name, p)
			}
		}
	}
	if got := res.Panels["mod.control.c1"]["properties"].(map[string]any)["severity"]; got != "high" {
		t.Errorf("expected unstripped property to remain, got %v", got)
	}
}

func TestWriteStrippedJson_PanelError(t *testing.T) {
	s := testSnapshot()
	s.Panels["mod.control.zzz"] = failingPanel{}

	var buf bytes.Buffer
	if err := s.WriteStrippedJson(&buf, false); err == nil {
		t.Fatal("expected an error for a panel that cannot be marshalled")
	}
	if _, err := s.AsStrippedJson(true); err == nil {
		t.Fatal("expected an error for a panel that cannot be marshalled")
	}
}

// the rows of a SnapshotRowsPanel are written one at a time: the whole panel is never marshalled
func TestWriteStrippedJson_RowsPanelNotMarshalledWhole(t *testing.T) {
	rows := make([]any, 1000)
	for i := range rows {
		rows[i] = map[string]any{"reason": fmt.Sprintf("row %d", i), "status": "ok"}
	}
	panel := &mapRowsPanel{head: map[string]any{"name": "mod.control.c", "data": map[string]any{"columns": []any{}}}, rows: rows}
	s := &SteampipeSnapshot{SchemaVersion: "20240607", Panels: map[string]SnapshotPanel{"mod.control.c": panel}}
	want, err := legacyAsStrippedJson(s, true)
	if err != nil {
		t.Fatal(err)
	}

	panel.failFullMarshal = true
	got, err := s.AsStrippedJson(true)
	if err != nil {
		t.Fatalf("AsStrippedJson: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("output differs from legacy\n got: %s\nwant: %s", got, want)
	}
}

func TestWriteStrippedJson_RowsPanelErrors(t *testing.T) {
	cases := map[string]struct {
		panel SnapshotPanel
		want  string
	}{
		"EachRow fails": {
			panel: &mapRowsPanel{head: map[string]any{"data": map[string]any{}}, rows: []any{"r1"}, err: errors.New("rows unavailable")},
			want:  "rows unavailable",
		},
		"row cannot be marshalled": {
			panel: &mapRowsPanel{head: map[string]any{"data": map[string]any{}}, rows: []any{"r1", make(chan int)}},
			want:  "unsupported type",
		},
		"head has rows": {
			panel: &mapRowsPanel{head: map[string]any{"data": map[string]any{"rows": []any{}}}},
			want:  "panel without rows has data rows",
		},
		"head is not an object": {
			panel: &mapRowsPanel{},
			want:  "must be a JSON object",
		},
		"data is not an object": {
			panel: &mapRowsPanel{head: map[string]any{"data": []any{1}}},
			want:  "panel data must be a JSON object or null",
		},
		"head cannot be marshalled": {
			panel: structRowsPanel{&testPanel{Error: errors.New("x"), Properties: map[string]any{"bad": make(chan int)}}},
			want:  "unsupported type",
		},
	}
	for name, tc := range cases {
		for _, indent := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/indent=%v", name, indent), func(t *testing.T) {
				s := testSnapshot()
				s.Panels["mod.control.zzz"] = tc.panel
				var buf bytes.Buffer
				err := s.WriteStrippedJson(&buf, indent)
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("expected an error containing %q, got %v", tc.want, err)
				}
			})
		}
	}
}
