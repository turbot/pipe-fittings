package modconfig

import (
	"fmt"
	"strings"
)

// AnonymousBlockName generates a unique name for an anonymous HCL block.
//
// The naming convention is: {sanitizedParentName}_anonymous_{blockType}_{childIndex}
//
// This function must be used by both eager and lazy loading paths to ensure
// consistent panel/resource naming regardless of loading strategy.
//
// Parameters:
//   - parentName: The unqualified name of the parent resource (e.g., "dashboard.my_dashboard")
//   - blockType: The type of the anonymous block (e.g., "container", "chart", "table")
//   - childIndex: The zero-based index of this child among siblings of the same type
//
// Example:
//
//	AnonymousBlockName("dashboard.my_dashboard", "container", 0)
//	// Returns: "dashboard_my_dashboard_anonymous_container_0"
func AnonymousBlockName(parentName, blockType string, childIndex int) string {
	sanitizedParentName := strings.ReplaceAll(parentName, ".", "_")
	return fmt.Sprintf("%s_anonymous_%s_%d", sanitizedParentName, blockType, childIndex)
}
