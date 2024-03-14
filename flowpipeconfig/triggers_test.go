package flowpipeconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFpTriggersFile(t *testing.T) {

	assert := assert.New(t)

	_, diags := LoadFlowpipeTriggerFile("./test/flowpipe.fptriggers")
	assert.Equal(0, len(diags))
}
