package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
)

func XTestQueryTriggerParse(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	_, triggers, err := load_mod.LoadPipelines(ctx, "./pipelines/query_trigger.fp")
	assert.Nil(err, "error found")

	scheduleTrigger := triggers["local.trigger.schedule.my_hourly_trigger"]
	if scheduleTrigger == nil {
		assert.Fail("my_hourly_trigger trigger not found")
		return
	}

	st, ok := scheduleTrigger.Config.(*modconfig.TriggerSchedule)
	if !ok {
		assert.Fail("my_hourly_trigger trigger is not a schedule trigger")
		return
	}

	assert.Equal("5 * * * *", st.Schedule)
}
