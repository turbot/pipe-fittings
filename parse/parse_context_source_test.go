package parse

import (
	"testing"
	"unsafe"

	"github.com/hashicorp/hcl/v2"
)

func TestParseContextSourceDefinition_MatchesGetSourceDefinition(t *testing.T) {
	fileData := map[string][]byte{
		"a.pp":     []byte("line 1\nline 2\nline 3\nline 4\nline 5\n"),
		"b.pp":     []byte("control \"c\" {\n  sql = \"select 1\"\n}"),
		"crlf.pp":  []byte("line 1\r\nline 2\r\n"),
		"empty.pp": {},
	}
	p := NewParseContext(".")
	p.SetDecodeContent(&hcl.BodyContent{}, fileData)

	rng := func(file string, start, end int) hcl.Range {
		return hcl.Range{Filename: file, Start: hcl.Pos{Line: start}, End: hcl.Pos{Line: end}}
	}
	ranges := []hcl.Range{
		rng("a.pp", 1, 1), rng("a.pp", 2, 4), rng("a.pp", 5, 5), rng("a.pp", 0, 2),
		rng("a.pp", 4, 9), rng("a.pp", 9, 12), rng("b.pp", 1, 3), rng("b.pp", 2, 2),
		rng("crlf.pp", 1, 2), rng("empty.pp", 1, 1), rng("missing.pp", 1, 1),
	}
	for _, r := range ranges {
		want := getSourceDefinition(r, fileData)
		if got := p.sourceDefinition(r); got != want {
			t.Errorf("%s %d-%d: got %q, want %q", r.Filename, r.Start.Line, r.End.Line, got, want)
		}
	}
}

// source definitions from the same file must share the file's memory, not copy it
func TestParseContextSourceDefinition_SharesFileMemory(t *testing.T) {
	content := []byte("block \"a\" {\n}\nblock \"b\" {\n  x = 1\n}\n")
	p := NewParseContext(".")
	p.SetDecodeContent(&hcl.BodyContent{}, map[string][]byte{"f.pp": content})

	a := p.sourceDefinition(hcl.Range{Filename: "f.pp", Start: hcl.Pos{Line: 1}, End: hcl.Pos{Line: 2}})
	b := p.sourceDefinition(hcl.Range{Filename: "f.pp", Start: hcl.Pos{Line: 3}, End: hcl.Pos{Line: 5}})
	if a != "block \"a\" {\n}" || b != "block \"b\" {\n  x = 1\n}" {
		t.Fatalf("unexpected source definitions %q %q", a, b)
	}

	file := p.fileSource["f.pp"]
	start := uintptr(unsafe.Pointer(unsafe.StringData(file)))
	end := start + uintptr(len(file))
	for _, s := range []string{a, b} {
		ptr := uintptr(unsafe.Pointer(unsafe.StringData(s)))
		if ptr < start || ptr >= end {
			t.Fatalf("source definition %q does not share the file's memory", s)
		}
	}

	// new decode content must not return source from the previous files
	p.SetDecodeContent(&hcl.BodyContent{}, map[string][]byte{"f.pp": []byte("other\n")})
	if got := p.sourceDefinition(hcl.Range{Filename: "f.pp", Start: hcl.Pos{Line: 1}, End: hcl.Pos{Line: 1}}); got != "other" {
		t.Fatalf("expected source from the new file content, got %q", got)
	}
}
