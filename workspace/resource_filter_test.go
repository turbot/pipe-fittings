package workspace

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/v2/modconfig"
	"github.com/turbot/pipe-fittings/v2/printers"
	"github.com/zclconf/go-cty/cty"
)

// MockHclResource implements modconfig.HclResource for testing
type MockHclResource struct {
	tags map[string]string
}

func (m *MockHclResource) GetTags() map[string]string {
	return m.tags
}

func (m *MockHclResource) GetShowData() *printers.RowData {
	data := printers.NewRowData(
		printers.NewFieldValue("tags", m.tags),
		printers.NewFieldValue("severity", "high"),
	)
	return data
}

// Implement other required methods with empty implementations
func (m *MockHclResource) Name() string                                   { return "test" }
func (m *MockHclResource) GetTitle() string                               { return "Test" }
func (m *MockHclResource) GetDescription() string                         { return "" }
func (m *MockHclResource) GetDocumentation() string                       { return "" }
func (m *MockHclResource) GetDeclRange() *hcl.Range                       { return nil }
func (m *MockHclResource) GetBlockType() string                           { return "test" }
func (m *MockHclResource) GetHclResourceImpl() *modconfig.HclResourceImpl { return nil }
func (m *MockHclResource) OnDecoded(*hcl.Block, modconfig.ModResourcesProvider) hcl.Diagnostics {
	return nil
}
func (m *MockHclResource) Equals(modconfig.HclResource) bool              { return false }
func (m *MockHclResource) GetCtyValue() (cty.Value, error)                { return cty.Zero, nil }
func (m *MockHclResource) SetBase(modconfig.HclResource)                  {}
func (m *MockHclResource) GetBase() modconfig.HclResource                 { return nil }
func (m *MockHclResource) IsTopLevel() bool                               { return false }
func (m *MockHclResource) SetTopLevel(bool)                               {}
func (m *MockHclResource) GetUnqualifiedName() string                     { return "test" }
func (m *MockHclResource) GetShortName() string                           { return "test" }
func (m *MockHclResource) GetFullName() string                            { return "test" }
func (m *MockHclResource) GetNestedStructs() []modconfig.CtyValueProvider { return nil }
func (m *MockHclResource) GetListData() *printers.RowData                 { return printers.NewRowData() }

func TestSimplePropertyFilter(t *testing.T) {
	resource := &MockHclResource{
		tags: map[string]string{
			"service":  "Azure/ActiveDirectory",
			"cis_type": "automated",
		},
	}

	tests := []struct {
		name     string
		where    string
		expected bool
	}{
		{
			name:     "severity equals high",
			where:    "severity='high'",
			expected: true,
		},
		{
			name:     "severity equals low",
			where:    "severity='low'",
			expected: false,
		},
		{
			name:     "severity not equals low",
			where:    "severity!='low'",
			expected: true,
		},
		{
			name:     "cis_type equals automated (using JSON path)",
			where:    "tags->>'cis_type' = 'automated'",
			expected: true,
		},
		{
			name:     "cis_type equals manual (using JSON path)",
			where:    "tags->>'cis_type' = 'manual'",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ResourceFilter{Where: tt.where}
			predicate, err := filter.getPredicate()
			if err != nil {
				t.Fatalf("Failed to create predicate: %v", err)
			}

			result := predicate(resource)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for filter: %s", tt.expected, result, tt.where)
			}
		})
	}
}

func TestJSONPathFilter(t *testing.T) {
	resource := &MockHclResource{
		tags: map[string]string{
			"service":  "Azure/ActiveDirectory",
			"cis_type": "automated",
		},
	}

	tests := []struct {
		name     string
		where    string
		expected bool
	}{
		{
			name:     "service not equals Azure/ActiveDirectory",
			where:    "tags->>'service' != 'Azure/ActiveDirectory'",
			expected: false,
		},
		{
			name:     "service equals Azure/ActiveDirectory",
			where:    "tags->>'service' = 'Azure/ActiveDirectory'",
			expected: true,
		},
		{
			name:     "service not in list",
			where:    "tags->>'service' not in ('Azure/EntraID','Azure/ActiveDirectory')",
			expected: false,
		},
		{
			name:     "service in list",
			where:    "tags->>'service' in ('Azure/EntraID','Azure/ActiveDirectory')",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ResourceFilter{Where: tt.where}
			predicate, err := filter.getPredicate()
			if err != nil {
				t.Fatalf("Failed to create predicate: %v", err)
			}

			result := predicate(resource)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for filter: %s", tt.expected, result, tt.where)
			}
		})
	}
}
