package entity

type DataLabelObject struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type NumberValueObject struct {
	Value float32 `json:"value"`
	Step  float32 `json:"step"`
	Min   float32 `json:"min"`
	Max   float32 `json:"max"`
}

type ResourceObject struct {
	Min NumberValueObject `json:"min"`
	Max NumberValueObject `json:"max"`
}

type PathObject struct {
	Default      string `json:"default"`
	Fixed        string `json:"fixed"`
	Customizable bool   `json:"customizable"`
}

type LimitsIntData struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type LimitsFloatData struct {
	Min float32 `json:"min"`
	Max float32 `json:"max"`
}

type ResourcesDataObject struct {
	Cpu    LimitsFloatData `json:"cpu"`
	Memory LimitsFloatData `json:"memory"`
}

type ApplicationData struct {
	Name            string              `json:"name"`
	RootPath        string              `json:"rootPath"`
	HealthCheckPath string              `json:"healthCheckPath"`
	Resources       ResourcesDataObject `json:"resources"`
	Port            int                 `json:"port"`
}

type IngressData struct {
	CustomHost     string `json:"customHost"`
	CustomPath     string `json:"customPath"`
	Authentication bool   `json:"authentication"`
}
