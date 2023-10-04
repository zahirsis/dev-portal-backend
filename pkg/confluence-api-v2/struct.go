package confluenceapiv2

type Version struct {
	Message string `json:"message"`
	Number  int    `json:"number"`
}

type BodyExtended struct {
	Representation string `json:"representation"`
	Value          string `json:"value"`
}

type Links struct {
	Webui  string `json:"webui"`
	Editui string `json:"editui"`
	Tinyui string `json:"tinyui"`
}

type ListLinks struct {
	Next string `json:"next"`
}
