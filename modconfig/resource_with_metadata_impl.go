package modconfig

import "github.com/hashicorp/hcl/v2"

type ResourceWithMetadataImpl struct {
	ResourceMetadata
	// required to allow partial decoding
	ResourceWithMetadataImplRemain hcl.Body             `hcl:",remain" json:"-"`
	References                     []*ResourceReference `json:"references,omitempty"`

	anonymous bool
}

// GetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) GetMetadata() *ResourceMetadata {
	return &b.ResourceMetadata
}

// SetMetadata implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) SetMetadata(metadata *ResourceMetadata) {
	if metadata != nil {
		b.ResourceMetadata = *metadata
		// set anonymous property on metadata
		b.ResourceMetadata.Anonymous = b.anonymous
	}
}

// SetAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) SetAnonymous(block *hcl.Block) {
	b.anonymous = len(block.Labels) == 0
}

// IsAnonymous implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) IsAnonymous() bool {
	return b.anonymous
}

// AddReference implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) AddReference(ref *ResourceReference) {
	b.References = append(b.References, ref)
}

// GetReferences implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) GetReferences() []*ResourceReference {
	return b.References
}

// GetRemain implements ResourceWithMetadata
func (b *ResourceWithMetadataImpl) GetResourceWithMetadataRemain() hcl.Body {
	return b.ResourceWithMetadataImplRemain
}
