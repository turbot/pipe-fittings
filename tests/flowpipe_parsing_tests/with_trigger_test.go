package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
)

func TestPipelineWithTrigger(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	pipelines, triggers, err := load_mod.LoadPipelines(ctx, "./pipelines/with_trigger.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.simple_with_trigger"] == nil {
		assert.Fail("simple_with_trigger pipeline not found")
		return
	}

	echoStep := pipelines["local.pipeline.simple_with_trigger"].GetStep("transform.simple_echo")
	if echoStep == nil {
		assert.Fail("transform.simple_echo step not found")
		return
	}

	dependsOn := echoStep.GetDependsOn()
	assert.Equal(len(dependsOn), 0)

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

	triggerWithArgs := triggers["local.trigger.schedule.trigger_with_args"]
	if triggerWithArgs == nil {
		assert.Fail("trigger_with_args trigger not found")
		return
	}

	twa, ok := triggerWithArgs.Config.(*modconfig.TriggerSchedule)
	if !ok {
		assert.Fail("trigger_with_args trigger is not a schedule trigger")
		return
	}

	assert.NotNil(twa, "trigger_with_args trigger is nil")

	// assert.Equal("one", triggerWithArgs.Args["param_one"])
	// assert.Equal(2, triggerWithArgs.Args["param_two_int"])

	queryTrigger := triggers["local.trigger.query.query_trigger"]
	if queryTrigger == nil {
		assert.Fail("query_trigger trigger not found")
		return
	}

	qt, ok := queryTrigger.Config.(*modconfig.TriggerQuery)
	if !ok {
		assert.Fail("query_trigger trigger is not a query trigger")
		return
	}

	assert.Equal("access_key_id", qt.PrimaryKey)
	assert.Len(qt.Events, 1)
	assert.Equal("insert", qt.Events[0])
	// assert.Equal("one", queryTrigger.Args["param_one"])
	// assert.Equal(2, queryTrigger.Args["param_two_int"])
	assert.Contains(qt.Sql, "where create_date < now() - interval")

	httpTriggerWithArgs := triggers["local.trigger.http.trigger_with_args"]
	if httpTriggerWithArgs == nil {
		assert.Fail("trigger_with_args trigger not found")
		return
	}

	_, ok = httpTriggerWithArgs.Config.(*modconfig.TriggerHttp)
	if !ok {
		assert.Fail("trigger_with_args trigger is not a schedule trigger")
		return
	}

}

func TestPipelineWithTriggerSelf(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	_, _, err := load_mod.LoadPipelines(ctx, "./pipelines/with_trigger_self.fp")
	assert.Nil(err, "error found")
}
