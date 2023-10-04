package service

type PageList struct {
	Id       string
	ParentId string
	Title    string
	Link     string
}

type WikiApiService interface {
	CreatePage(title, space, parent string, content []byte) (url string, err error)
	ListSubPages(space, parent string) ([]*PageList, error)
	UpdatePage(Id string, content []byte, updateMessage string) error
}
