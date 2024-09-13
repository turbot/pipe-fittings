package modconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestCtyTypeSerialisation(t *testing.T) {
	assert := assert.New(t)

	ctyType := cty.String

	assert.Equal(ctyType, cty.String)

	// JSON Serialise
	jsonBytes, err := ctyType.MarshalJSON()
	assert.Nil(err)
	assert.Equal([]byte(`"string"`), jsonBytes)

	ctyType = cty.List(cty.String)
	jsonBytes, err = ctyType.MarshalJSON()
	assert.Nil(err)
	assert.Equal(`["list","string"]`, string(jsonBytes))

	// capsule type can't be serialied :(

	// ctyType = cty.Capsule("aws", reflect.TypeOf(AwsConnection{}))
	// jsonBytes, err = ctyType.MarshalJSON()
	// assert.Nil(err)
	// assert.Equal(`["capsule","aws","modconfig.AwsConnection"]`, string(jsonBytes))

}
