package modconfig

import "github.com/turbot/pipe-fittings/schema"

type ILoop interface {
	GetType() string
}

func GetLoopDefn(stepType string) ILoop {
	switch stepType {
	case schema.BlockTypePipelineStepEcho:
		return &LoopEchoStep{}
	case schema.BlockTypePipelineStepHttp:
		return &LoopHttpStep{}
	}
	return nil
}

type LoopEchoStep struct {
	If      bool    `json:"if" hcl:"if"`
	Numeric *int    `json:"numeric,omitempty" hcl:"numeric,optional"`
	Text    *string `json:"text,omitempty" hcl:"text,optional"`
}

func (*LoopEchoStep) GetType() string {
	return schema.BlockTypePipelineStepEcho
}

type LoopHttpStep struct {
	If  bool    `json:"if" hcl:"if"`
	Url *string `json:"url,omitempty" hcl:"url,optional"`
}

func (*LoopHttpStep) GetType() string {
	return schema.BlockTypePipelineStepHttp
}
