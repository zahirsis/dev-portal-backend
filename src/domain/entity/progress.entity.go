package entity

import "time"

type ProgressEntity interface {
	Time() time.Time
	Step() string
	Message() string
	Kind() string
	IsNode() bool
	ToStruct() Progress
}

type progressEntity struct {
	time    time.Time
	step    string
	message string
	kind    string
	Node    bool
}

type Progress struct {
	Time    time.Time `json:"time"`
	Step    string    `json:"step"`
	Message string    `json:"message"`
	Kind    string    `json:"type"`
	Node    bool      `json:"node"`
}

func NewProgressEntity(progress Progress) ProgressEntity {
	return &progressEntity{
		time:    progress.Time,
		step:    progress.Step,
		message: progress.Message,
		kind:    progress.Kind,
		Node:    progress.Node,
	}
}

func (t *progressEntity) Time() time.Time {
	return t.time
}

func (t *progressEntity) Step() string {
	return t.step
}

func (t *progressEntity) Message() string {
	return t.message
}

func (t *progressEntity) Kind() string {
	return t.kind
}

func (t *progressEntity) IsNode() bool {
	return t.Node
}

func (t *progressEntity) ToStruct() Progress {
	return Progress{
		Time:    t.time,
		Step:    t.step,
		Message: t.message,
		Kind:    t.kind,
		Node:    t.Node,
	}
}
