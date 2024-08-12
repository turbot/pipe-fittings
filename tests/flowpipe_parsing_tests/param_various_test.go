package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
)

func TestParamVarious(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := load_mod.LoadPipelines(context.TODO(), "./pipelines/param_various.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.param_various"]
	if pipeline == nil {
		assert.Fail("Pipeline not found")
		return
	}

	for _, param := range pipeline.Params {

		if param.Name == "foo" {
			assert.Equal("any", param.TypeHCLString) // type not speficied we assume dynamic
		} else if param.Name == "list_of_string" {
			assert.Equal("list(string)", param.TypeHCLString)
		} else if param.Name == "map_of_number" {
			assert.Equal("map(number)", param.TypeHCLString)
		} else if param.Name == "map_of_bool" {
			assert.Equal("map(bool)", param.TypeHCLString)
		} else if param.Name == "map_of_list_of_number" {
			assert.Equal("map(list(number))", param.TypeHCLString)
		} else if param.Name == "map_of_a_map_of_a_bool" {
			assert.Equal("map(map(bool))", param.TypeHCLString)
		} else if param.Name == "map_of_any" {
			assert.Equal("map(any)", param.TypeHCLString)
		} else if param.Name == "list_of_list_of_string" {
			assert.Equal("list(list(string))", param.TypeHCLString)
		} else if param.Name == "list_of_map_of_bool" {
			assert.Equal("list(map(bool))", param.TypeHCLString)
		} else if param.Name == "list_of_map_of_list_of_number" {
			assert.Equal("list(map(list(number)))", param.TypeHCLString)
		} else if param.Name == "list_of_map_of_list_of_string" {
			assert.Equal("list(map(list(string)))", param.TypeHCLString)
		} else {
			assert.Fail("Unknown param: ", param.Name)
		}
	}

}
