package funcs

import (
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var Base64URLDecodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0]
		s := str.AsString()
		sDec, err := base64.URLEncoding.DecodeString(s)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		if !utf8.Valid([]byte(sDec)) {
			//log.Printf("[DEBUG] the result of decoding the provided string is not valid UTF-8: %s", redactIfSensitive(sDec, strMarks))
			return cty.UnknownVal(cty.String), fmt.Errorf("the result of decoding the provided string is not valid UTF-8")
		}
		return cty.StringVal(string(sDec)), nil
	},
})