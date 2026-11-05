package parse

import (
	"maps"
	"slices"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/turbot/pipe-fittings/v2/modconfig"
)

const testModSource = `
locals {
  common = query.used_by_locals.sql
  tags   = { service = "AWS/S3" }
}

benchmark "top" {
  title    = "Top"
  children = [benchmark.child, control.direct, mymod.control.self_qualified]
}

benchmark "child" {
  children = [control.nested]
  tags     = local.tags
}

control "direct" {
  query = query.q_direct
  title = "Uses ${control.in_template.title} in a template"
}

control "nested" {
  sql = query.q_sql.sql
  param "p" {
    default = query.q_param.sql
  }
}

control "self_qualified" {
  query = query.q_dep_user
  description = aws.control.from_dep.description
}

control "in_template" {
  sql   = "select 1"
  title = var.title
}

control "self_input" {
  sql   = "select 1"
  title = self.input.x.value
}

control "unused" {
  query = query.q_unused
}

dashboard "unused" {
  card {
    query = query.q_unused
  }
}

query "q_direct" { sql = "select 1" }
query "q_sql" { sql = "select 1" }
query "q_param" { sql = "select 1" }
query "q_dep_user" { sql = "select 1" }
query "used_by_locals" { sql = "select 1" }
query "q_unused" { sql = "select 1" }
`

func testModContent(t *testing.T, source string) *hcl.BodyContent {
	t.Helper()
	file, diags := hclparse.NewParser().ParseHCL([]byte(source), "mod.pp")
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	content, diags := file.Body.Content(WorkspaceBlockSchema)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	return content
}

func TestNeededResources(t *testing.T) {
	content := testModContent(t, testModSource)

	needed, depRefs, ok := neededResources(content, "mymod", []string{"benchmark.top"}, nil)
	if !ok {
		t.Fatal("expected the needed resources to be known")
	}
	want := []string{
		"benchmark.child", "benchmark.top",
		"control.direct", "control.in_template", "control.nested", "control.self_qualified",
		"query.q_dep_user", "query.q_direct", "query.q_param", "query.q_sql", "query.used_by_locals",
	}
	if got := slices.Sorted(maps.Keys(needed)); !slices.Equal(got, want) {
		t.Fatalf("needed resources\n got: %v\nwant: %v", got, want)
	}
	if got := depRefs["aws"]; !slices.Equal(got, []string{"control.from_dep"}) {
		t.Fatalf("expected a reference to aws control.from_dep, got %v", depRefs)
	}
	if len(depRefs) != 1 {
		t.Fatalf("expected references to one other mod only, got %v", depRefs)
	}
}

func TestNeededResources_OptionalTargets(t *testing.T) {
	content := testModContent(t, testModSource)

	// an optional target that this mod does not define is not needed, and is not an error
	needed, _, ok := neededResources(content, "mymod", nil, []string{"control.self_input", "benchmark.in_another_mod"})
	if !ok {
		t.Fatal("expected the needed resources to be known")
	}
	want := []string{"control.self_input", "query.used_by_locals"}
	if got := slices.Sorted(maps.Keys(needed)); !slices.Equal(got, want) {
		t.Fatalf("needed resources\n got: %v\nwant: %v", got, want)
	}
}

func TestNeededResources_Unknown(t *testing.T) {
	cases := map[string]struct {
		source   string
		required []string
	}{
		"required target not defined": {
			source:   testModSource,
			required: []string{"benchmark.missing"},
		},
		// for example a query from a .sql file, or a block nested in a dashboard
		"reference to a resource that is not a top level block": {
			source:   `control "c" { query = query.from_sql_file }`,
			required: []string{"control.c"},
		},
		"reference from locals to a resource that is not a top level block": {
			source:   "locals {\n  q = query.missing\n}\ncontrol \"c\" {\n  sql = \"select 1\"\n}",
			required: []string{"control.c"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, ok := neededResources(testModContent(t, tc.source), "mymod", tc.required, nil); ok {
				t.Fatal("expected the needed resources to be unknown")
			}
		})
	}
}

func TestNeededResources_NotNativeSyntax(t *testing.T) {
	file, diags := hclparse.NewParser().ParseJSON([]byte(`{"control": {"c": {"sql": "select 1"}}}`), "mod.json")
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	content, diags := file.Body.Content(WorkspaceBlockSchema)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	if _, _, ok := neededResources(content, "mymod", []string{"control.c"}, nil); ok {
		t.Fatal("expected the needed resources of a JSON body to be unknown")
	}
}

func TestTraversalReference(t *testing.T) {
	cases := map[string]resourceReference{
		"control.c1":                {name: "control.c1"},
		"control.c1.title":          {name: "control.c1"},
		"query.q1.sql":              {name: "query.q1"},
		"mymod.benchmark.b1":        {name: "benchmark.b1"},
		"aws.control.c1.title":      {mod: "aws", name: "control.c1"},
		"var.region":                {},
		"local.tags":                {},
		"self.input.x.value":        {},
		"param.p":                   {},
		"aws.var.region":            {},
		"control.c1[0]":             {name: "control.c1"},
		"local.queries.query.inner": {},
		"local.control.c1":          {},
		"var.control.c1":            {},
	}
	for expr, want := range cases {
		parsed, diags := hclparse.NewParser().ParseHCL([]byte("x = "+expr), "test.pp")
		if diags.HasErrors() {
			t.Fatalf("%s: %v", expr, diags)
		}
		attrs, _ := parsed.Body.JustAttributes()
		traversals := attrs["x"].Expr.Variables()
		if len(traversals) != 1 {
			t.Fatalf("%s: expected one traversal, got %d", expr, len(traversals))
		}
		got, isResource := traversalReference(traversals[0], "mymod")
		if isResource != (want != resourceReference{}) || got != want {
			t.Errorf("%s: got %+v (resource %v), want %+v", expr, got, isResource, want)
		}
	}
}

func newTestModParseContext(t *testing.T, modName string, parent *ModParseContext, targets ...ResourceTarget) *ModParseContext {
	t.Helper()
	m, err := NewModParseContext(nil, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m.CurrentMod = &modconfig.Mod{}
	m.CurrentMod.ShortName = modName
	m.ParentParseCtx = parent
	m.SetResourceTargets(targets...)
	return m
}

func TestSetResourceFilter(t *testing.T) {
	content := testModContent(t, testModSource)

	// a target in another mod only: no resources of this mod are needed
	m := newTestModParseContext(t, "mymod", nil, ResourceTarget{Mod: "aws", Name: "benchmark.all"})
	m.setResourceFilter(content)
	if !m.HasResourceTargets() || len(m.resourceFilter) != 1 {
		t.Fatalf("expected only the resource used by locals, got %v", m.resourceFilter)
	}

	// the resources that a parent mod references are needed
	parent := newTestModParseContext(t, "parent", nil, ResourceTarget{Mod: "parent", Name: "benchmark.p"})
	parent.dependencyResourceRefs = map[string][]string{"mymod": {"control.nested"}}
	child := newTestModParseContext(t, "mymod", parent, parent.resourceTargets...)
	child.setResourceFilter(content)
	for _, name := range []string{"control.nested", "query.q_sql", "query.q_param"} {
		if _, ok := child.resourceFilter[name]; !ok {
			t.Errorf("expected %s to be needed, got %v", name, child.resourceFilter)
		}
	}
	if _, ok := child.resourceFilter["control.unused"]; ok {
		t.Errorf("control.unused is not needed")
	}

	// if the needed resources are unknown, everything is decoded, here and in dependency mods
	unknown := newTestModParseContext(t, "mymod", nil, ResourceTarget{Mod: "mymod", Name: "benchmark.missing"})
	unknown.setResourceFilter(content)
	if unknown.HasResourceTargets() || unknown.resourceFilter != nil {
		t.Fatal("expected no resource filter")
	}

	// no targets: no filter
	none := newTestModParseContext(t, "mymod", nil)
	none.setResourceFilter(content)
	if none.HasResourceTargets() || none.resourceFilter != nil {
		t.Fatal("expected no resource filter without targets")
	}
}

func TestShouldIncludeBlock_ResourceFilter(t *testing.T) {
	content := testModContent(t, testModSource)
	m := newTestModParseContext(t, "mymod", nil, ResourceTarget{Mod: "mymod", Name: "control.direct"})
	m.setResourceFilter(content)

	var included []string
	for _, block := range m.filterBlocks(content.Blocks) {
		name, isResource := resourceBlockName(block)
		if !isResource {
			name = block.Type
		}
		included = append(included, name)
	}
	slices.Sort(included)
	want := []string{"control.direct", "control.in_template", "locals", "query.q_direct", "query.used_by_locals"}
	if !slices.Equal(included, want) {
		t.Fatalf("included blocks\n got: %v\nwant: %v", included, want)
	}
}
