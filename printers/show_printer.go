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

type ShowPrinter[T any] struct {
}

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
	au := aurora.NewAurora(opts.ColorEnabled)

	var b strings.Builder

	/* we print primitive types as follows
	<TitleFormat(Title)>:<padding><value>
	*/
	// calc the padding
	// the padding is such that the value is aligned with the longest title
	var maxTitleLength int
	for _, c := range row.Columns {
		if len(c) > maxTitleLength {
			maxTitleLength = len(c)
		}
	}
	// add 2 for the colon and space
	maxTitleLength += 2

	for idx, columnName := range row.Columns {
		fieldOpts := row.Opts[columnName]

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
			fieldString = p.renderKeyValue(maxTitleLength, au, columnName, row.Cells[idx], fieldOpts, opts)
		}

		// if there is anything in the field string, add it to the builder, with indent
		if len(fieldString) > 0 {
			b.WriteString(indentString)
			b.WriteString(fieldString)
		}
	}

	return b.String(), nil
}

func (p ShowPrinter[T]) renderKeyValue(maxTitleLength int, au aurora.Aurora, columnName string, columnValue any, fieldOpts FieldRenderOptions, opts sanitize.RenderOptions) string {
	// key
	padFormat := fmt.Sprintf("%%-%ds", maxTitleLength)
	keyStr := fmt.Sprintf(padFormat, au.Blue(fmt.Sprintf("%s:", columnName)))

	// value
	var valstr string

	// if there is a value render func, use it to render the value
	if fieldOpts.RenderValueFunc != nil {
		valstr = fieldOpts.RenderValueFunc(opts) + "\n"
	} else {
		// ok manually render the value
		var err error
		valstr, err = p.renderValue(columnValue, opts)
		if err != nil {
			return ""
		}
	}
	// now combine
	return fmt.Sprintf("%s%s", keyStr, valstr)
}

func (p ShowPrinter[T]) renderValue(value any, opts sanitize.RenderOptions) (string, error) {
	// todo what do we do with nil values? exclude from source data?
	if helpers.IsNil(value) {
		return "\n", nil
	}
	val := reflect.ValueOf(value)

	// if the element is showable, call show on it
	if s := AsShowable(value); s != nil {
		childOpts := opts.Clone()
		childOpts.Indent += 2
		return p.render(s, childOpts)
	}

	var valStr string

	switch val.Kind() {
	case reflect.Slice:
		sliceString, err := p.showSlice(val, opts)
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
			vt = dereferencePointer(vt)
			valStr = fmt.Sprintf("%v\n", vt)
		}
	}

	return valStr, nil
}

func (p ShowPrinter[T]) showSlice(val reflect.Value, opts sanitize.RenderOptions) (string, error) {
	var b strings.Builder

	for i := 0; i < val.Len(); i++ {
		// Retrieve each element in the slice
		//elem := dereferencePointer(val.Index(i).Interface())
		elem := val.Index(i).Interface()

		valStr, err := p.renderValue(elem, opts)
		if err != nil {
			return "", err
		}
		b.WriteString(valStr)

	}
	return b.String(), nil
}
