package connection

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/zclconf/go-cty/cty"
	"maps"
)

// no hcl tags needed apart from pipes block - this is a manually populated

type ConnectionImpl struct {
	Pipes *PipesConnectionMetadata `json:"pipes,omitempty" cty:"pipes" hcl:"pipes,block"`

	ShortName string `json:"short_name" cty:"short_name"`
	FullName  string `json:"full_name,omitempty" cty:"name"`
	// DeclRange uses the hclhelpers.Range type which reimplements hcl.Range with custom serialisation
	DeclRange hclhelpers.Range `json:"decl_range,omitempty" cty:"decl_range"`
	// cache ttl in seconds
	Ttl int `json:"ttl,omitempty" cty:"ttl" hcl:"ttl,optional"`
}

func NewConnectionImpl(connectionType, shortName string, declRange hcl.Range) ConnectionImpl {
	return ConnectionImpl{
		ShortName: shortName,
		FullName:  fmt.Sprintf("%s.%s", connectionType, shortName),
		DeclRange: hclhelpers.NewRange(declRange),
		Ttl:       -1,
	}
}

func (c *ConnectionImpl) Name() string {
	return c.FullName
}

func (c *ConnectionImpl) GetShortName() string {
	return c.ShortName
}

func (c *ConnectionImpl) GetConnectionType() string {
	panic("method GetConnectionType must be implemented by concrete connection type")
}

func (c *ConnectionImpl) GetConnectionImpl() *ConnectionImpl {
	return c
}

func (c *ConnectionImpl) GetTtl() int {
	return c.Ttl
}

func (c *ConnectionImpl) SetTtl(ttl int) {
	c.Ttl = ttl
}

func (c *ConnectionImpl) Equals(other *ConnectionImpl) bool {
	if c.ShortName != other.ShortName {
		return false
	}
	if c.FullName != other.FullName {
		return false
	}
	if !c.DeclRange.Equals(other.DeclRange) {
		return false
	}

	if c.Ttl != other.Ttl {
		return false
	}
	return true
}

// CustomType implements custom_type.CustomType interface
func (c *ConnectionImpl) CustomType() {
}

// LateBinding implements the LateBinding interface, marking this as a type whose value is not known until runtime
func (c *ConnectionImpl) LateBinding() {
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
	mergedValueMap["type"] = cty.StringVal(connection.GetConnectionType())
	mergedValueMap["resource_type"] = cty.StringVal("connection." + connection.GetConnectionType())
	return cty.ObjectVal(mergedValueMap), nil
}
