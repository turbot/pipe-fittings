package parse

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type coerceValueTest struct {
	title    string
	resource modconfig.ResourceWithParam
	input    map[string]string
	expected map[string]interface{}
}

type resourceWithParams struct {
	Params []modconfig.PipelineParam
}

func (r *resourceWithParams) GetParam(name string) *modconfig.PipelineParam {
	for _, p := range r.Params {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func (r *resourceWithParams) GetParams() []modconfig.PipelineParam {
	return r.Params
}

var coerceValueTests = []coerceValueTest{
	{
		title: "Coerce string value",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_string",
					Type: cty.String,
				},
			},
		},
		input: map[string]string{
			"param_string": "val_one",
		},
		expected: map[string]interface{}{
			"param_string": "val_one",
		},
	},
	{
		title: "Coerce int value",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_int",
					Type: cty.Number,
				},
			},
		},
		input: map[string]string{
			"param_int": "123",
		},
		expected: map[string]interface{}{
			"param_int": 123,
		},
	},
	{
		title: "Coerce bool value",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_bool",
					Type: cty.Bool,
				},
			},
		},
		input: map[string]string{
			"param_bool": "true",
		},
		expected: map[string]interface{}{
			"param_bool": true,
		},
	},
	{
		title: "Coerce connection",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Capsule("aws", reflect.TypeOf(modconfig.AwsConnection{})),
				},
			},
		},
		input: map[string]string{
			"param_connection": "connection.aws.default",
		},
		expected: map[string]interface{}{
			"param_connection": map[string]interface{}{
				"name":          "default",
				"type":          "aws",
				"resource_type": "connection",
				"temporary":     true,
			},
		},
	},
	{
		title: "Coerce notifier",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_notifier",
					Type: cty.Capsule("notifier", reflect.TypeOf(&modconfig.NotifierImpl{})),
				},
			},
		},
		input: map[string]string{
			"param_notifier": "notifier.slack",
		},
		expected: map[string]interface{}{
			"param_notifier": map[string]interface{}{
				"name":          "slack",
				"resource_type": "notifier",
			},
		},
	},
	{
		title: "list of connections",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(modconfig.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default,connection.aws.example]",
		},
		expected: map[string]interface{}{
			"param_connection": []interface{}{
				map[string]interface{}{
					"name":          "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
				map[string]interface{}{
					"name":          "example",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
			},
		},
	},
	{
		title: "list of connection but just one",
		resource: &resourceWithParams{
			Params: []modconfig.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(modconfig.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default]",
		},
		expected: map[string]interface{}{
			"param_connection": []interface{}{
				map[string]interface{}{
					"name":          "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
			},
		},
	},
	// {
	// 	title: "map of connections",
	// 	resource: &resourceWithParams{
	// 		Params: []modconfig.PipelineParam{
	// 			{
	// 				Name: "param_connection",
	// 				Type: cty.Map(cty.Capsule("aws", reflect.TypeOf(modconfig.AwsConnection{}))),
	// 			},
	// 		},
	// 	},
	// 	input: map[string]string{
	// 		"param_connection": "{default=connection.aws.default,example=connection.aws.example}",
	// 	},
	// 	expected: map[string]interface{}{
	// 		"param_connection": map[string]interface{}{
	// 			"default": map[string]interface{}{
	// 				"name":          "default",
	// 				"type":          "aws",
	// 				"resource_type": "connection",
	// 				"temporary":     true,
	// 			},
	// 			"example": map[string]interface{}{
	// 				"name":          "example",
	// 				"type":          "aws",
	// 				"resource_type": "connection",
	// 				"temporary":     true,
	// 			},
	// 		},
	// 	},
	// },
}

func TestCoerceCustomValue(tm *testing.T) {

	variables := map[string]cty.Value{
		"connection": cty.ObjectVal(map[string]cty.Value{
			"aws": cty.ObjectVal(map[string]cty.Value{
				"default": cty.ObjectVal(map[string]cty.Value{
					"name":          cty.StringVal("default"),
					"type":          cty.StringVal("aws"),
					"temporary":     cty.BoolVal(true),
					"resource_type": cty.StringVal("connection"),
				}),
				"example": cty.ObjectVal(map[string]cty.Value{
					"name":          cty.StringVal("example"),
					"type":          cty.StringVal("aws"),
					"temporary":     cty.BoolVal(true),
					"resource_type": cty.StringVal("connection"),
				}),
			}),
		}),
		"notifier": cty.ObjectVal(map[string]cty.Value{
			"slack": cty.ObjectVal(map[string]cty.Value{
				"name":          cty.StringVal("slack"),
				"resource_type": cty.StringVal("notifier"),
			}),
			"default": cty.ObjectVal(map[string]cty.Value{
				"name":          cty.StringVal("default"),
				"resource_type": cty.StringVal("notifier"),
			}),
		}),
	}
	evalCtx := &hcl.EvalContext{
		Variables: variables,
	}

	for _, tc := range coerceValueTests {
		tm.Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)

			// pass nil evalCtx so it will not validate the actual connection/notifier against the
			// config
			result, err := CoerceParams(tc.resource, tc.input, evalCtx)
			if len(err) > 0 {
				assert.Fail("Error while coercing value", "error", err)
				return
			}

			assert.True(reflect.DeepEqual(tc.expected, result))
		})
	}
}
