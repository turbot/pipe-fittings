package printers

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/sanitize"
	"reflect"
	"strings"
	"time"
)

type Showable interface {
	GetShowData() *Table
}
type Listable interface {
	GetListData() *Table
}

func Show(resource Showable, opts sanitize.RenderOptions) (string, error) {
	data := resource.GetShowData()
	au := aurora.NewAurora(opts.ColorEnabled)
	if len(data.Rows) != 1 {
		return "", fmt.Errorf("expected 1 row, got %d", len(data.Rows))
	}
	row := data.Rows[0]
	if len(data.Columns) != len(row.Cells) {
		return "", fmt.Errorf("expected %d columns, got %d", len(data.Columns), len(data.Rows[0].Cells))
	}

	var b strings.Builder

	/* we print primitive types as follows
	<TitleFormat(Title)>:<padding><value>
	*/
	// calc the padding
	// the padding is such that the value is aligned with the longest title
	var maxTitleLength int
	for _, c := range data.Columns {
		if len(c.Name) > maxTitleLength {
			maxTitleLength = len(c.Name)
		}
	}
	// add 2 for the colon and space
	maxTitleLength += 2

	for idx, c := range data.Columns {
		fieldOpts := data.FieldOpts[c.Name]

		// if render opts or this field specify an indent, apply it
		globalIndent := opts.Indent
		fieldIndent := fieldOpts.Indent
		var indentString = strings.Repeat(" ", globalIndent+fieldIndent)

		var fieldString string

		// if there is a key-value render func, use it to render the full key and value
		if fieldOpts.RenderKeyValueFunc != nil {
			fieldString = fieldOpts.RenderKeyValueFunc(opts)
			// is this returned anything, add a newline
			if len(fieldString) > 0 {
				fieldString += "\n"
			}
		} else {
			fieldString = renderKeyValue(maxTitleLength, au, c, row.Cells[idx], fieldOpts, opts)
		}

		// if there is anything in the field string, add it to the builder, with indent
		if len(fieldString) > 0 {
			b.WriteString(indentString)
			b.WriteString(fieldString)
		}
	}

	return b.String(), nil
}

func renderKeyValue(maxTitleLength int, au aurora.Aurora, c TableColumnDefinition, columnValue any, fieldOpts FieldRenderOptions, opts sanitize.RenderOptions) string {
	// key
	padFormat := fmt.Sprintf("%%-%ds", maxTitleLength)
	keyStr := fmt.Sprintf(padFormat, au.Blue(fmt.Sprintf("%s:", c.Name)))

	// value
	var valstr string

	// if there is a value render func, use it to render the value
	if fieldOpts.RenderValueFunc != nil {
		valstr = fieldOpts.RenderValueFunc(opts) + "\n"
	} else {
		// ok manually render the value
		var err error
		valstr, err = renderValue(columnValue, opts)
		if err != nil {
			return ""
		}
	}
	// now combine
	return fmt.Sprintf("%s%s", keyStr, valstr)
}

func renderValue(value any, opts sanitize.RenderOptions) (string, error) {
	val := reflect.ValueOf(value)

	// if the element is showable, call show on it
	if s, ok := value.(Showable); ok {
		childOpts := opts.Clone()
		childOpts.Indent += 2
		return Show(s, childOpts)
	}

	var valStr string

	switch val.Kind() {
	case reflect.Slice:
		sliceString, err := showSlice(val, opts)
		if err != nil {
			return "", err
		}
		// put a newline BEFORE the slice string
		valStr = "\n" + sliceString

		// TODO  handle map, struct
	// case reflect.Map:
	// case reflect.Struct:

	default:
		switch vt := value.(type) {
		case time.Time:
			valStr = fmt.Sprintf("%v\n", vt.Format(time.RFC3339))
		default:
			// todo dereference ptr?????
			valStr = fmt.Sprintf("%v\n", vt)
		}
	}

	return valStr, nil
}

func showSlice(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	for i := 0; i < val.Len(); i++ {
		// Retrieve each element in the slice
		elem := dereferencePointer(val.Index(i).Interface())

		valStr, err := renderValue(elem, opts)
		if err != nil {
			return "", err
		}
		b.WriteString(valStr)

	}
	return b.String(), nil
}

//// addBullet takes a multiline string.
//// It adds "- " to the start of the first line and indents all other lines to align with the bullet.
//func addBullet(s string) string {
//	lines := strings.Split(s, "\n")
//
//	// Process first line with bullet
//	if len(lines) > 0 {
//		lines[0] = "- " + lines[0]
//	}
//
//	// Process remaining lines
//	indent := strings.Repeat(" ", len("- "))
//	for i := 1; i < len(lines); i++ {
//		if len(lines[i]) > 0 {
//			lines[i] = indent + lines[i]
//		}
//	}
//
//	return strings.Join(lines, "\n")
//
//}

func renderField(key string, value any, level int, au aurora.Aurora) string {
	if !helpers.IsNil(value) {
		return fmt.Sprintf("%s%s\n", au.Blue(key+":").Bold(), value)
	}
	return ""
}
