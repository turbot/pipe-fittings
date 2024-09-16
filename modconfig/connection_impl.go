package modconfig

import "github.com/hashicorp/hcl/v2"

type ConnectionImpl struct {
	HclResourceImpl

	// required to allow partial decoding
	HclResourceRemain hcl.Body `hcl:",remain" json:"-"`

	Type string `json:"type" cty:"type" hcl:"type,label"`
}

func (c *ConnectionImpl) GetUnqualifiedName() string {
	return c.HclResourceImpl.UnqualifiedName
}

func (c *ConnectionImpl) SetHclResourceImpl(hclResourceImpl HclResourceImpl) {
	c.HclResourceImpl = hclResourceImpl
}
