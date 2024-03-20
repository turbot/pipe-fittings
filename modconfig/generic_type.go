package modconfig

import (
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"strings"
)

// GenericTypeToBlockType converts a resource type generic param into a block type
// NOTE special case handling for dashboard items
func GenericTypeToBlockType[T ModTreeItem]() string {
	var resourceType string
	var empty T
	switch any(empty).(type) {
	case Variable:
		resourceType = schema.AttributeVar
	case *DashboardCard:
		resourceType = schema.BlockTypeCard
	case *DashboardChart:
		resourceType = schema.BlockTypeChart
	case *DashboardContainer:
		resourceType = schema.BlockTypeContainer
	case *DashboardFlow:
		resourceType = schema.BlockTypeFlow
	case *DashboardGraph:
		resourceType = schema.BlockTypeGraph
	case *DashboardHierarchy:
		resourceType = schema.BlockTypeHierarchy
	case *DashboardImage:
		resourceType = schema.BlockTypeImage
	case *DashboardInput:
		resourceType = schema.BlockTypeInput
	case *DashboardTable:
		resourceType = schema.BlockTypeTable
	case *DashboardText:
		resourceType = schema.BlockTypeText
	default:
		resourceType = strings.ToLower(utils.GetGenericTypeName[T]())
	}
	return resourceType
}
