package hclhelpers

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
)

// Range represents a span of characters between two positions in a source file.
// This is a direct re-implementation of hcl.Range, allowing us to control JSON serialization
type Range struct {
	// Filename is the name of the file into which this range's positions point.
	Filename string `json:"filename,omitempty" cty:"filename"`

	// Start and End represent the bounds of this range. Start is inclusive and End is exclusive.
	Start Pos `json:"start,omitempty" cty:"start"`
	End   Pos `json:"end,omitempty" cty:"end"`
}

func (r Range) HclRange() hcl.Range {
	return hcl.Range{
		Filename: r.Filename,
		Start:    r.Start.HclPos(),
		End:      r.End.HclPos(),
	}
}

func (r Range) HclRangePointer() *hcl.Range {
	return &hcl.Range{
		Filename: r.Filename,
		Start:    r.Start.HclPos(),
		End:      r.End.HclPos(),
	}
}

func (r Range) Equals(declRange Range) bool {
	return r.Filename == declRange.Filename && r.Start == declRange.Start && r.End == declRange.End
}

func NewRange(sourceRange hcl.Range) Range {
	return Range{
		Filename: sourceRange.Filename,
		Start:    NewPos(sourceRange.Start),
		End:      NewPos(sourceRange.End),
	}
}

// strin
// string
func (r Range) String() string {
	return fmt.Sprintf("Range{Filename: %s, Start: %s, End: %s}", r.Filename, r.Start, r.End)
}

// Pos represents a single position in a source file
// This is a direct re-implementation of hcl.Pos, allowing us to control JSON serialization
type Pos struct {
	Line   int `json:"line" cty:"line"`
	Column int `json:"column" cty:"column"`
	Byte   int `json:"byte" cty:"byte"`
}

func (p Pos) HclPos() hcl.Pos {
	return hcl.Pos{
		Line:   p.Line,
		Column: p.Column,
		Byte:   p.Byte,
	}
}

// string
func (p Pos) String() string {
	return fmt.Sprintf("Pos{Line: %d, Column: %d, Byte: %d}", p.Line, p.Column, p.Byte)
}

func NewPos(sourcePos hcl.Pos) Pos {
	return Pos{
		Line:   sourcePos.Line,
		Column: sourcePos.Column,
		Byte:   sourcePos.Byte,
	}
}
