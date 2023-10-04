package confluenceapiv2

import (
	"net/url"
	"strconv"
)

type PageSortOrder string

const (
	TitleAsc         PageSortOrder = "title"
	TitleDesc        PageSortOrder = "-title"
	IdAsc            PageSortOrder = "id"
	IdDesc           PageSortOrder = "-id"
	CreatedDateAsc   PageSortOrder = "created-date"
	CreatedDateDesc  PageSortOrder = "-created-date"
	ModifiedDateAsc  PageSortOrder = "modified-date"
	ModifiedDateDesc PageSortOrder = "-modified-date"
)

type Status string

const (
	Current  Status = "current"
	Archived Status = "archived"
	Deleted  Status = "deleted"
	Trashed  Status = "trashed"
	Draft    Status = "draft"
)

type GetPageByIdQuery struct {
	BodyFormat string `json:"body-format,omitempty"`
	GetDraft   bool   `json:"get-draft,omitempty"`
	Version    int    `json:"version,omitempty"`
}

type GetPagesInSpaceQuery struct {
	Limit  int           `json:"limit,omitempty"`
	Cursor string        `json:"cursor,omitempty"`
	Sort   PageSortOrder `json:"sort,omitempty"`
	Status Status        `json:"status,omitempty"`
}

type PageBody struct {
	Storage        *BodyExtended `json:"storage"`
	AtlasDocFormat *BodyExtended `json:"atlas_doc_format"`
	View           *BodyExtended `json:"view"`
}

type Page struct {
	Id         string    `json:"id,omitempty"`
	Status     Status    `json:"status,omitempty"`
	Title      string    `json:"title,omitempty"`
	SpaceId    string    `json:"spaceId,omitempty"`
	ParentId   string    `json:"parentId,omitempty"`
	ParentType string    `json:"parentType,omitempty"`
	Position   int       `json:"position,omitempty"`
	AuthorId   string    `json:"authorId,omitempty"`
	CreatedAt  string    `json:"createdAt,omitempty"`
	Version    *Version  `json:"version,omitempty"`
	Body       *PageBody `json:"body,omitempty"`
	Links      *Links    `json:"_links,omitempty"`
}

type Pages struct {
	Results []*Page    `json:"results"`
	Links   *ListLinks `json:"_links"`
}

// getPageIDEndpoint creates the correct api endpoint by given id
func (a *API) getPageIDEndpoint(id string) (*url.URL, error) {
	return url.ParseRequestURI(a.endPoint.String() + "/pages/" + id)
}

// getPageEndpoint creates the correct api endpoint
func (a *API) getSpacePagesEndpoint(id string) (*url.URL, error) {
	return url.ParseRequestURI(a.endPoint.String() + "/spaces/" + id + "/pages")
}

// getPageEndpoint creates the correct api endpoint
func (a *API) getPageEndpoint() (*url.URL, error) {
	return url.ParseRequestURI(a.endPoint.String() + "/pages/")
}

// getPageChildEndpoint creates the correct api endpoint by given id and type
func (a *API) getPageChildEndpoint(id string, t string) (*url.URL, error) {
	return url.ParseRequestURI(a.endPoint.String() + "/pages/" + id + "/child/" + t)
}

// getPageGenericEndpoint creates the correct api endpoint by given id and type
func (a *API) getPageGenericEndpoint(id string, t string) (*url.URL, error) {
	return url.ParseRequestURI(a.endPoint.String() + "/pages/" + id + "/" + t)
}

// GetPageByID queries pages by id
func (a *API) GetPageByID(id string, query *GetPageByIdQuery) (*Page, error) {
	ep, err := a.getPageIDEndpoint(id)
	if err != nil {
		return nil, err
	}
	ep.RawQuery = addPageByIdQueryParams(query).Encode()
	return a.SendPageRequest(ep, "GET", nil)
}

// GetPagesInSpace queries pages in a space
func (a *API) GetPagesInSpace(space string, query *GetPagesInSpaceQuery) (*Pages, error) {
	ep, err := a.getSpacePagesEndpoint(space)
	if err != nil {
		return nil, err
	}
	ep.RawQuery = addPagesInSpaceQueryParams(query).Encode()
	return a.SendPagesRequest(ep, "GET", nil)
}

// UpdatePage updates a page
func (a *API) UpdatePage(page *Page) (*Page, error) {
	ep, err := a.getPageIDEndpoint(page.Id)
	if err != nil {
		return nil, err
	}
	return a.SendPageRequest(ep, "PUT", page)
}

// CreatePage creates a page
func (a *API) CreatePage(page *Page) (*Page, error) {
	ep, err := a.getPageEndpoint()
	if err != nil {
		return nil, err
	}
	return a.SendPageRequest(ep, "POST", page)
}

// addPageByIdQueryParams adds the defined query parameters
func addPageByIdQueryParams(query *GetPageByIdQuery) *url.Values {

	data := url.Values{}

	if query.BodyFormat != "" {
		data.Set("body-format", query.BodyFormat)
	}
	if !query.GetDraft {
		data.Set("get-draft", "false")
	} else {
		data.Set("get-draft", "true")
	}
	//get specific version
	if query.Version != 0 {
		data.Set("version", strconv.Itoa(query.Version))
	}
	return &data
}

// addPagesInSpaceQueryParams adds the defined query parameters
func addPagesInSpaceQueryParams(query *GetPagesInSpaceQuery) *url.Values {

	data := url.Values{}

	if query.Limit != 0 {
		data.Set("limit", strconv.Itoa(query.Limit))
	}
	if query.Cursor != "" {
		data.Set("cursor", query.Cursor)
	}
	if query.Sort != "" {
		data.Set("sort", string(query.Sort))
	}
	if query.Status != "" {
		data.Set("status", string(query.Status))
	}
	return &data
}
