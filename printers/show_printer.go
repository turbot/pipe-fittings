package printers

import (
	"context"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/sanitize"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
)

type ShowPrinter[T any] struct{}

func NewShowPrinter[T any]() (*ShowPrinter[T], error) {
	return &ShowPrinter[T]{}, nil
}

func (p ShowPrinter[T]) PrintResource(_ context.Context, r PrintableResource[T], writer io.Writer) error {
	items := r.GetItems()
	if len(items) != 1 {
		return fmt.Errorf("expected exactly one item, got %d", len(items))
	}

	enableColor := viper.GetString(constants.ArgOutput) == constants.OutputFormatPretty
	renderOpts := sanitize.RenderOptions{ColorEnabled: enableColor}

	showable, ok := any(items[0]).(Showable)
	if !ok {
		// this is a bug
		return fmt.Errorf("expected a showable item, got %T", items[0])
	}

	str, err := p.render(showable, renderOpts)
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(str)); err != nil {
		return fmt.Errorf("error printing resource")
	}
	return nil
}

func (p ShowPrinter[T]) render(resource Showable, opts sanitize.RenderOptions) (string, error) {
	row := resource.GetShowData()
	return p.renderShowData(row, opts)
}

func (p ShowPrinter[T]) renderShowData(row *RowData, opts sanitize.RenderOptions) (string, error) {
	au := aurora.NewAurora(opts.ColorEnabled)

	var b strings.Builder

	/* we print fields types as follows

		<TitleFormat(Title)>:<padding><value>

	the padding is such that the value is aligned with the longest title	*/
	maxTitleLength := p.getMaxTitleLength(row)

	for _, columnName := range row.Columns {
		fieldVal := row.Fields[columnName]

		displayName := row.displayNameMap[columnName]
		// if render renderOpts or this field specify an indent, apply it
		indent := opts.Indent
		if opts.IsList && !fieldVal.renderOpts.listOpts.isKey {
			indent += 2
		}

		var indentString = strings.Repeat(" ", indent)
		var fieldString string

		// if there is a key-value render func, use it to render the full key and value
		if opts.IsList && fieldVal.renderOpts.listOpts.listKeyRenderFunc != nil {
			fieldString = fieldVal.renderOpts.listOpts.listKeyRenderFunc(opts)

			// if the string is empty, skip
			if len(fieldString) == 0 {
				continue
			}
			fieldString += "\n"

		} else {
			fieldString = p.renderKeyValue(fieldVal, displayName, maxTitleLength, au, opts)
		}

		// if there is anything in the field string, add it to the builder, with indent
		if len(fieldString) > 0 {
			b.WriteString(indentString)
			b.WriteString(fieldString)
		}
	}

	// if the string is empty, just return it
	if b.Len() == 0 {
		return "", nil
	}
	// if this is NOT a list and not top level data, put a newline BEFORE the string
	var valStr string
	if !opts.IsList && opts.Indent > 0 {
		valStr = "\n" + b.String()
	} else {
		valStr = b.String()
	}

	return valStr, nil
}

func (p ShowPrinter[T]) getMaxTitleLength(row *RowData) int {
	var maxTitleLength int
	for _, c := range row.Columns {
		if len(c) > maxTitleLength {
			maxTitleLength = len(c)
		}
	}
	// add 2 for the colon and space
	maxTitleLength += 2
	return maxTitleLength
}

func (p ShowPrinter[T]) renderKeyValue(fieldVal FieldValue, columnName string, maxTitleLength int, au aurora.Aurora, opts sanitize.RenderOptions) string {
	// key
	padFormat := fmt.Sprintf("%%-%ds", maxTitleLength)
	keyStr := fmt.Sprintf(padFormat, au.Blue(fmt.Sprintf("%s:", columnName)))

	// value
	var valStr string

	// if there is a value render func, use it to render the value
	if fieldVal.renderOpts.renderValueFunc != nil {
		valStr = fieldVal.renderOpts.renderValueFunc(opts)
		// skip empty values
		if valStr == "" {
			return ""
		}
		// add newline
		valStr += "\n"
	} else {
		// ok manually render the value
		var err error
		valStr, err = p.renderValue(fieldVal.Value, opts)
		if err != nil {
			return ""
		}
		if valStr == "" {
			return ""
		}
	}

	// TODO KAI CAN WE ALWAYS ADD NEWLINE _ CHECK WITH FLOWPIPE
	// now combine
	return fmt.Sprintf("%s%s", keyStr, valStr)
}

// render the given value
// called by renderKeyValue and renderSlice
func (p ShowPrinter[T]) renderValue(value any, opts sanitize.RenderOptions) (string, error) {
	if helpers.IsNil(value) {
		return "", nil
	}
	val := reflect.ValueOf(value)

	// if the element is showable, call show on it
	if s := AsShowable(value); s != nil {
		childOpts := opts.Clone()
		// if this is NOT a list item, indent the struct by 2
		// (list items are already indented)
		if !opts.IsList {
			childOpts.Indent += 2
		}
		return p.render(s, childOpts)
	}

	var valStr string
	var err error
	switch val.Kind() {
	case reflect.Slice:
		valStr, err = p.renderSlice(val, opts)
		if err != nil {
			return "", err
		}
	case reflect.Map:
		valStr, err = p.renderMap(val, opts)
		if err != nil {
			return "", err
		}
	case reflect.Struct:
		valStr, err = p.renderStruct(val, opts)
		if err != nil {
			return "", err
		}
	default:
		valStr = p.renderPrimitive(value, opts)
	}

	return valStr, nil
}

func (p ShowPrinter[T]) renderPrimitive(value any, opts sanitize.RenderOptions) string {
	// is this is a list item, apply the indent
	var valStr string

	switch vt := value.(type) {
	case time.Time:
		valStr = fmt.Sprintf("%v", vt.Format(time.RFC3339))
	default:
		vt = dereferencePointer(vt)
		valStr = fmt.Sprintf("%v", vt)
	}
	// if the value is non empty, add indent (if required) and  newline
	if valStr != "" {
		var indent string
		if opts.IsList {
			indent = strings.Repeat(" ", opts.Indent)
		}

		valStr = indent + valStr + "\n"
	}

	return valStr
}

func (p ShowPrinter[T]) renderSlice(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	// clone the opts and increment the indent
	childOpts := opts.Clone()
	childOpts.IsList = true
	// indent the slice by 2
	childOpts.Indent += 2

	for i := 0; i < val.Len(); i++ {
		// Retrieve each element in the slice
		//elem := dereferencePointer(val.Index(i).Interface())
		elem := val.Index(i).Interface()

		valStr, err := p.renderValue(elem, childOpts)
		if err != nil {
			return "", err
		}
		//// add bullet
		//valStr = p.addBullet(valStr)
		b.WriteString(valStr)

	}

	// if the string is empty, just return it
	if b.Len() == 0 {
		return "", nil
	}
	// if this is NOT a list, put a newline BEFORE the string
	var valStr string
	if !opts.IsList {
		valStr = "\n" + b.String()
	} else {
		valStr = b.String()
	}

	return valStr, nil
}

func (p ShowPrinter[T]) renderStruct(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	// clone the opts and increment the indent
	childOpts := opts.Clone()
	// NOTE:  clear the IsList flag - this will make  renderShowData
	// put newlines before and after each struct in a list of structs
	childOpts.IsList = false
	childOpts.Indent += 2

	asMap := helpers.StructToMap(val.Interface())
	keys := helpers.SortedMapKeys(asMap)

	var fieldValues []FieldValue
	for _, k := range keys {
		v := asMap[k]
		fieldValues = append(fieldValues, NewFieldValue(k, v))
	}

	// convert to RowData and render
	showData := NewRowData(fieldValues...)
	str, err := p.renderShowData(showData, childOpts)
	if err != nil {
		return "", err

	}
	b.WriteString(str)

	return b.String(), nil
}

func (p ShowPrinter[T]) renderMap(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	// clone the opts and increment the indent
	childOpts := opts.Clone()
	// NOTE:  clear the IsList flag - this will make  renderShowData
	// put newlines before and after each struct in a list of maps
	childOpts.IsList = false

	if !opts.IsList {
		// indent the struct by 2
		childOpts.Indent += 2
	}

	// convert to map[string]any
	var asMap = map[string]any{}
	for _, key := range val.MapKeys() {
		// Ensure the key is a string
		keyStr := fmt.Sprintf("%v", key)
		asMap[keyStr] = val.MapIndex(key).Interface()
	}

	// now sort keys and create field values
	var fieldValues []FieldValue

	keys := helpers.SortedMapKeys(asMap)
	for _, k := range keys {
		v := asMap[k]
		fieldValues = append(fieldValues, NewFieldValue(k, v))
	}
	// convert to RowData and render
	showData := NewRowData(fieldValues...)
	str, err := p.renderShowData(showData, childOpts)
	if err != nil {
		return "", err
	}
	b.WriteString(str)

	return b.String(), nil
}

//
//func (p ShowPrinter[T]) addBullet(str string) string {
//	lines := strings.Split(str, "\n")
//	for i, line := range lines {
//		if len(line) == 0 {
//			continue
//		}
//		if i == 0 {
//			lines[i] = " -" + line
//		} else {
//			lines[i] = "  " + line
//		}
//	}
//	return strings.Join(lines, "\n")
//}
