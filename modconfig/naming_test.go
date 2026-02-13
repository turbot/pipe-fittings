package modconfig

import "testing"

func TestAnonymousBlockName(t *testing.T) {
	tests := []struct {
		name        string
		parentName  string
		blockType   string
		childIndex  int
		expected    string
	}{
		{
			name:       "dashboard with container",
			parentName: "dashboard.my_dashboard",
			blockType:  "container",
			childIndex: 0,
			expected:   "dashboard_my_dashboard_anonymous_container_0",
		},
		{
			name:       "dashboard with multiple containers",
			parentName: "dashboard.sibling_containers_report",
			blockType:  "container",
			childIndex: 2,
			expected:   "dashboard_sibling_containers_report_anonymous_container_2",
		},
		{
			name:       "container with chart",
			parentName: "container.dashboard_my_dashboard_anonymous_container_0",
			blockType:  "chart",
			childIndex: 0,
			expected:   "container_dashboard_my_dashboard_anonymous_container_0_anonymous_chart_0",
		},
		{
			name:       "container with text",
			parentName: "container.my_container",
			blockType:  "text",
			childIndex: 1,
			expected:   "container_my_container_anonymous_text_1",
		},
		{
			name:       "dashboard with table",
			parentName: "dashboard.testing_card_blocks",
			blockType:  "table",
			childIndex: 0,
			expected:   "dashboard_testing_card_blocks_anonymous_table_0",
		},
		{
			name:       "zero index",
			parentName: "dashboard.test",
			blockType:  "card",
			childIndex: 0,
			expected:   "dashboard_test_anonymous_card_0",
		},
		{
			name:       "high index",
			parentName: "dashboard.test",
			blockType:  "input",
			childIndex: 99,
			expected:   "dashboard_test_anonymous_input_99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnonymousBlockName(tt.parentName, tt.blockType, tt.childIndex)
			if result != tt.expected {
				t.Errorf("AnonymousBlockName(%q, %q, %d) = %q, want %q",
					tt.parentName, tt.blockType, tt.childIndex, result, tt.expected)
			}
		})
	}
}
