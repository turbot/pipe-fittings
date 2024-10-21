package search_path

import "fmt"

type SearchPathConfig struct {
	SearchPath       []string
	SearchPathPrefix []string
}

func (c SearchPathConfig) Empty() bool {
	return len(c.SearchPath) == 0 && len(c.SearchPathPrefix) == 0
}

func (c SearchPathConfig) String() string {
	if c.Empty() {
		return ""
	}
	if len(c.SearchPath) > 0 {
		return fmt.Sprintf("search_path=%v", c.SearchPath)
	}
	return fmt.Sprintf("search_path_prefix=%v", c.SearchPathPrefix)
}
