package plugin

import "github.com/hashicorp/hcl/v2"

// Range represents a span of characters between two positions in a source file.
// This is a direct re-implementation of hcl.Range, allowing us to control JSON serialization
type Range struct {
	// Filename is the name of the file into which this range's positions point.
	Filename string `json:"filename,omitempty"`

	// Start and End represent the bounds of this range. Start is inclusive and End is exclusive.
	Start Pos `json:"start,omitempty"`
	End   Pos `json:"end,omitempty"`
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
	Line   int `json:"line"`
	Column int `json:"column"`
	Byte   int `json:"byte"`
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
