package confluence

import (
	"fmt"
	"github.com/zahirsis/dev-portal-backend/config"
	confluenceapiv2 "github.com/zahirsis/dev-portal-backend/pkg/confluence-api-v2"
	"github.com/zahirsis/dev-portal-backend/src/domain/service"
	"github.com/zahirsis/dev-portal-backend/src/pkg/logger"
	"net/url"
)

type confluenceService struct {
	config *config.Config
	logger logger.Logger
	api    *confluenceapiv2.API
}

func NewConfluenceService(config *config.Config, logger logger.Logger, api *confluenceapiv2.API) service.WikiApiService {
	return &confluenceService{config, logger, api}
}

func (c *confluenceService) CreatePage(title, space, parent string, content []byte) (url string, err error) {
	c.logger.Debug(fmt.Sprintf("Creating page %s", title))
	wc, err := c.api.CreatePage(&confluenceapiv2.Page{
		Status:   "current",
		Title:    title,
		SpaceId:  space,
		ParentId: parent,
		Body: &confluenceapiv2.PageBody{
			Storage: &confluenceapiv2.BodyExtended{Value: string(content), Representation: "storage"},
		},
	})
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error creating page %s: %s", title, err.Error()))
		return "", err
	}
	return fmt.Sprintf("%s/wiki%s", c.config.WikiConfig.BaseUrl, wc.Links.Webui), nil
}

func (c *confluenceService) ListSubPages(space, parent string) ([]*service.PageList, error) {
	finished := false
	limit := 250
	cursor := ""
	var pages []*service.PageList
	for !finished {
		p, err := c.api.GetPagesInSpace(space, &confluenceapiv2.GetPagesInSpaceQuery{
			Limit:  limit,
			Cursor: cursor,
			Sort:   confluenceapiv2.TitleAsc,
			Status: confluenceapiv2.Current,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range p.Results {
			if v.ParentId == parent {
				pages = append(pages, &service.PageList{Id: v.Id, Title: v.Title, Link: fmt.Sprintf("/wiki%s", v.Links.Webui)})
			}
		}
		if p.Links.Next == "" {
			finished = true
		} else {
			u, err := url.Parse(p.Links.Next)
			if err != nil {
				c.logger.Error(fmt.Sprintf("Error parsing url %s: %s", p.Links.Next, err.Error()))
				return nil, err
			}
			cursor = u.Query().Get("cursor")
		}
	}
	return pages, nil
}

func (c *confluenceService) UpdatePage(Id string, content []byte, updateMessage string) error {
	page, err := c.api.GetPageByID(Id, &confluenceapiv2.GetPageByIdQuery{})
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error getting page %s: %s", Id, err.Error()))
		return err
	}
	_, err = c.api.UpdatePage(&confluenceapiv2.Page{
		Id:      page.Id,
		SpaceId: page.SpaceId,
		Status:  page.Status,
		Title:   page.Title,
		Body: &confluenceapiv2.PageBody{
			Storage: &confluenceapiv2.BodyExtended{
				Representation: "storage",
				Value:          string(content),
			},
		},
		Version: &confluenceapiv2.Version{
			Message: updateMessage,
			Number:  page.Version.Number + 1,
		},
	})
	if err != nil {
		c.logger.Error(fmt.Sprintf("Error updating page %s: %s", Id, err.Error()))
		return err
	}
	return nil
}
