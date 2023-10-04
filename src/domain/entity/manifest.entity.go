package entity

type Manifest struct {
	Code  string       `json:"code"`
	Label string       `json:"label"`
	Type  ManifestType `json:"type"`
	Dir   string       `json:"dir,omitempty"`
}
