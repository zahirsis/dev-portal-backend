package entity

type SquadEntity interface {
	Code() string
	Label() string
}

type squadEntity struct {
	label DataLabelObject
}

func NewSquadEntity(code string, label string) SquadEntity {
	return &squadEntity{
		DataLabelObject{
			Code:  code,
			Label: label,
		},
	}
}

func (t *squadEntity) Code() string {
	return t.label.Code
}

func (t *squadEntity) Label() string {
	return t.label.Label
}
