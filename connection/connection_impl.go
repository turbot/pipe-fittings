package connection

import (
	"fmt"
	"maps"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/zclconf/go-cty/cty"
)

// no hcl tags needed - this is a manually populated
type ConnectionImpl struct {
	Type      string `json:"type" cty:"type"`
	ShortName string `json:"short_name" cty:"short_name"`
	// DeclRange uses the hclhelpers.Range type which reimplements hcl.Range with custom serialisation
	DeclRange hclhelpers.Range `json:"decl_range,omitempty" cty:"decl_range"`
}

func NewConnectionImpl(block *hcl.Block) ConnectionImpl {
	var blockType, shortName string
	// handle the case where there are no labels - this is expected as an mepty connection object may be created
	if len(block.Labels) > 0 {
		blockType = block.Labels[0]
	}
	if len(block.Labels) > 1 {
		shortName = block.Labels[1]
	}
	return ConnectionImpl{
		Type:      blockType,
		ShortName: shortName,
		DeclRange: hclhelpers.NewRange(block.DefRange),
	}
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
