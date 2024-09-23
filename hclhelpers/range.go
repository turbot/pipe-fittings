package hclhelpers

import "github.com/hashicorp/hcl/v2"

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

func NewRange(sourceRange hcl.Range) Range {
	return Range{
		Filename: sourceRange.Filename,
		Start:    NewPos(sourceRange.Start),
		End:      NewPos(sourceRange.End),
	}
}

// Pos represents a single position in a source file
// This is a direct re-implementation of hcl.Pos, allowing us to control JSON serialization
type Pos struct {
	Line   int `json:"line" cty:"line"`
	Column int `json:"column" cty:"column"`
	Byte   int `json:"byte" cty:"byte"`
}

func (r Pos) HclPos() hcl.Pos {
	return hcl.Pos{
		Line:   r.Line,
		Column: r.Column,
		Byte:   r.Byte,
	}
}

func NewPos(sourcePos hcl.Pos) Pos {
	return Pos{
		Line:   sourcePos.Line,
		Column: sourcePos.Column,
		Byte:   sourcePos.Byte,
	}
}
