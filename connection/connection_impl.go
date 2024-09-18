package connection

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/zclconf/go-cty/cty"
	"maps"
)

type ConnectionImpl struct {
	Type      string    `json:"type" cty:"type" hcl:"type,label"`
	ShortName string    `json:"short_name" cty:"short_name" hcl:"short_name,label"`
	DeclRange hcl.Range `json:"decl_range,omitempty" hcl:"-"`
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
	baseValueMap := baseCtyValue.AsValueMap()

	// copy into base, overriding base properties with derived properties if where there are clashes
	maps.Copy(baseValueMap, valueMap)

	// we will return base
	baseValueMap["env"] = cty.ObjectVal(connection.GetEnv())

	return cty.ObjectVal(baseValueMap), nil
}
