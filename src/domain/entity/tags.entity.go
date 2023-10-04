package entity

type Tag struct {
	Key   *string `json:"key" yaml:"key"`
	Value *string `json:"value" yaml:"value"`
}

func NewTag(key string, value string) *Tag {
	return &Tag{
		Key:   &key,
		Value: &value,
	}
}

func DefaultTags(e SetupCiCdEntity) []*Tag {
	return []*Tag{
		NewTag("app", e.ApplicationName()),
		NewTag("squad", e.Squad().Code()),
		NewTag("cloud", "true"),
		NewTag("automated-setup", "true"),
	}
}
