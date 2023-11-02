package modconfig

import "github.com/turbot/pipe-fittings/schema"

type LoopDefn interface {
	ShouldRun() bool
	GetType() string
	UpdateInput(input Input) (Input, error)
}

func GetLoopDefn(stepType string) LoopDefn {
	switch stepType {
	case schema.BlockTypePipelineStepEcho:
		return &LoopEchoStep{}
	case schema.BlockTypePipelineStepHttp:
		return &LoopHttpStep{}
	}
	return nil
}

type LoopEchoStep struct {
	If      bool    `json:"if" hcl:"if" cty:"if"`
	Numeric *int    `json:"numeric,omitempty" hcl:"numeric,optional" cty:"numeric"`
	Text    *string `json:"text,omitempty" hcl:"text,optional" cty:"text"`
}

func (l *LoopEchoStep) UpdateInput(input Input) (Input, error) {
	if l.Numeric != nil {
		input["numeric"] = *l.Numeric
	}
	if l.Text != nil {
		input["text"] = *l.Text
	}
	return input, nil
}

func (l *LoopEchoStep) ShouldRun() bool {
	return l.If
}

func (*LoopEchoStep) GetType() string {
	return schema.BlockTypePipelineStepEcho
}

type LoopHttpStep struct {
	If  bool    `json:"if" hcl:"if" cty:"if"`
	Url *string `json:"url,omitempty" hcl:"url,optional" cty:"url"`
}

func (l *LoopHttpStep) ShouldRun() bool {
	return l.If
}

func (l *LoopHttpStep) UpdateInput(input Input) (Input, error) {
	if l.Url != nil {
		input["url"] = *l.Url
	}
	return input, nil
}

func (*LoopHttpStep) GetType() string {
	return schema.BlockTypePipelineStepHttp
}
