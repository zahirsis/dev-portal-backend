package confluenceapiv2

import (
	"net/http"
	"net/url"
)

type API struct {
	endPoint        *url.URL
	Client          *http.Client
	username, token string
}
