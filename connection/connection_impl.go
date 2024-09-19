package connection

import (
	"fmt"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/zclconf/go-cty/cty"
	"maps"
)

// no hcl tags needed - this is a manually populated
type ConnectionImpl struct {
	Type      string `json:"type" cty:"type"`
	ShortName string `json:"short_name" cty:"short_name"`
	// DeclRange uses the hclhelpers.Range type which reimplements hcl.Range with custom serialisation
	DeclRange hclhelpers.Range `json:"decl_range,omitempty" cty:"decl_range"`
}

func (c *ConnectionImpl) Name() string {
	return fmt.Sprintf("%s.%s", c.Type, c.ShortName)
}

func (c *ConnectionImpl) GetShortName() string {
	return c.ShortName
}

func (c *ConnectionImpl) GetConnectionType() string {
	return c.Type
}

func (c *ConnectionImpl) GetConnectionImpl() ConnectionImpl {
	return *c
}

func ctyValueForConnection(connection PipelingConnection) (cty.Value, error) {
	ctyValue, err := cty_helpers.GetCtyValue(connection)
	if err != nil {
		return cty.NilVal, err
	}
	impl := connection.GetConnectionImpl()
	baseCtyValue, err := cty_helpers.GetCtyValue(impl)
	if err != nil {
		return cty.NilVal, err
	}

	valueMap := ctyValue.AsValueMap()
	mergedValueMap := baseCtyValue.AsValueMap()

	// copy into mergedValueMap, overriding base properties with derived properties if where there are clashes
	// we will return mergedValueMap
	maps.Copy(mergedValueMap, valueMap)

	mergedValueMap["env"] = cty.ObjectVal(connection.GetEnv())
	return cty.ObjectVal(mergedValueMap), nil
}
