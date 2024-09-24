package querydisplay

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/queryresult"
	pqueryresult "github.com/turbot/pipe-fittings/queryresult"
	"github.com/turbot/pipe-fittings/utils"
)

const snapshotTemplate = `
{
  "schema_version": "{{ .SchemaVersion }}",
  "panels": {
    {{- range $panel := .Panels }}
    "{{ $panel.Dashboard }}": {
      "dashboard": "{{ $panel.Dashboard }}",
      "name": "{{ $panel.Name }}",
      "panel_type": "{{ $panel.PanelType }}",
      "source_definition": "{{ $panel.SourceDefinition }}",
      "status": "{{ $panel.Status }}",
      "title": "{{ $panel.Title }}"
    },
    "custom.table.results": {
      "dashboard": "{{ $panel.Dashboard }}",
      "name": "custom.table.results",
      "panel_type": "table",
      "source_definition": "{{ $panel.SourceDefinition }}",
      "status": "{{ $panel.Status }}",
      "sql": "{{ $panel.SQL }}",
      "properties": {
        "name": "results"
      },
      {{- if $panel.Data }}
      "data": {
        "columns": [
          {{- range $i, $col := $panel.Data.Columns }}
          {
            "name": "{{ $col.Name }}",
            "data_type": "{{ $col.DataType }}",
            "original_name": "{{ $col.OriginalName }}"
          }{{ if lt (add1 $i) (len $panel.Data.Columns) }},{{ end }}
          {{- end }}
        ],
        "rows": [
          {{- range $rowIndex, $row := $panel.Data.Rows }}
          {
            {{- $rowLen := len $row }}
            {{- $currentIndex := 0 }}
            {{- range $key, $value := $row }}
            "{{ $key }}": {{ $value }}{{ if lt (add1 $currentIndex) $rowLen }},{{ end }}
            {{- $currentIndex = add1 $currentIndex }}
            {{- end }}
          }{{ if lt (add1 $rowIndex) (len $panel.Data.Rows) }},{{ end }}
          {{- end }}
        ]
      }
      {{- end }}
    }
    {{- end }}
  },
  "inputs": {},
  "variables": {},
  "search_path": [
    {{- range $i, $path := .SearchPath }}
    "{{ $path }}"{{ if lt (add1 $i) (len $.SearchPath) }},{{ end }}
    {{- end }}
  ],
  "start_time": "{{ .StartTime }}",
  "end_time": "{{ .EndTime }}",
  "layout": {
    "name": "{{ .Layout.Name }}",
    "children": [
      {
        "name": "custom.table.results",
        "panel_type": "table"
      }
    ],
    "panel_type": "dashboard"
  }
}
`

// a snapshot in steampipe is generated using templating since we do not have a snapshot data type
// generateSnapshot function using Go templating
func generateSnapshot[T any](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery, searchPath []string, startTime time.Time) (bytes.Buffer, error) {
	var out bytes.Buffer
	endTime := time.Now()
	// Initialize the template
	tmpl, err := template.New("snapshot").Funcs(template.FuncMap{
		"add1": func(i int) int { return i + 1 },
	}).Parse(snapshotTemplate)
	if err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("unable to parse snapshot template: %w", err))
		return out, err
	}
	// Build the snapshot data (use the new getData function to retrieve data)
	snapshotData := map[string]any{
		"SchemaVersion": "20221222",
		"Panels":        getPanels[T](ctx, result, resolvedQuery),
		"SearchPath":    searchPath,
		"StartTime":     startTime.Format(time.RFC3339),
		"EndTime":       endTime.Format(time.RFC3339),
		"Layout":        getLayout[T](result, resolvedQuery),
	}
	// Render the template
	if err := tmpl.Execute(&out, snapshotData); err != nil {
		error_helpers.ShowError(ctx, fmt.Errorf("unable to execute snapshot template: %w", err))
		return out, err
	}
	return out, nil
}

type PanelData struct {
	Dashboard        string
	Name             string
	PanelType        string
	SourceDefinition string
	Title            string
	Status           string
	SQL              string
	Data             map[string]interface{}
}

func getPanels[T any](ctx context.Context, result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) map[string]PanelData {
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return nil
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Build panel data with proper fields
	return map[string]PanelData{
		"panelKey": {
			Dashboard:        dashboardName,
			Name:             dashboardName,
			PanelType:        "dashboard",
			SourceDefinition: "",
			Title:            fmt.Sprintf("Custom query [%s]", hash),
			Status:           "complete",
			SQL:              resolvedQuery.RawSQL,
			Data:             getData(ctx, result),
		},
	}
}

func getData[T any](ctx context.Context, result *queryresult.Result[T]) map[string]interface{} {
	jsonOutput := newJSONOutput[T]()
	// Ensure columns are being added
	if len(result.Cols) == 0 {
		error_helpers.ShowError(ctx, fmt.Errorf("no columns found in the result"))
	}
	// Add column definitions to the JSON output
	for _, col := range result.Cols {
		c := pqueryresult.ColumnDef{
			Name:         col.Name,
			OriginalName: col.OriginalName,
			DataType:     strings.ToUpper(col.DataType),
		}
		jsonOutput.Columns = append(jsonOutput.Columns, c)
	}
	// Define function to add each row to the JSON output
	rowFunc := func(row []interface{}, result *queryresult.Result[T]) {
		record := map[string]interface{}{}
		for idx, col := range result.Cols {
			value, _ := ParseJSONOutputColumnValue(row[idx], col)
			record[col.Name] = value
		}
		jsonOutput.Rows = append(jsonOutput.Rows, record)
	}
	// Call iterateResults and ensure rows are processed
	_, err := iterateResults(result, rowFunc)
	if err != nil {
		error_helpers.ShowError(ctx, err)
	}
	// Return the full data (including columns and rows)
	return map[string]interface{}{
		"Columns": jsonOutput.Columns,
		"Rows":    jsonOutput.Rows,
	}
}

func getLayout[T any](result *queryresult.Result[T], resolvedQuery *modconfig.ResolvedQuery) map[string]any {
	hash, err := utils.Base36Hash(resolvedQuery.RawSQL, 8)
	if err != nil {
		return nil
	}
	dashboardName := fmt.Sprintf("custom.dashboard.sql_%s", hash)
	// Define layout structure
	return map[string]any{
		"Name": dashboardName,
	}
}
