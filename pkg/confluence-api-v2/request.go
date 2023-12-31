package confluenceapiv2

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Request implements the basic Request function
func (a *API) Request(req *http.Request) ([]byte, error) {
	req.Header.Add("Accept", "application/json, */*")

	// only auth if we can auth
	if (a.username != "") || (a.token != "") {
		a.Auth(req)
	}

	Debug("====== Request ======")
	Debug(req)
	Debug("====== Request Body ======")
	if DebugFlag {
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(requestDump))
	}
	Debug("====== /Request Body ======")
	Debug("====== /Request ======")

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	Debug(fmt.Sprintf("====== Response Status Code: %d ======", resp.StatusCode))

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	Debug("====== Response Body ======")
	Debug(string(res))
	Debug("====== /Response Body ======")

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		return res, nil
	case http.StatusNoContent, http.StatusResetContent:
		return nil, nil
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed")
	case http.StatusServiceUnavailable:
		return nil, fmt.Errorf("service is not available: %s", resp.Status)
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error: %s", resp.Status)
	case http.StatusConflict:
		return nil, fmt.Errorf("conflict: %s", resp.Status)
	}

	return nil, fmt.Errorf("unknown response status: %s", resp.Status)
}

// SendPageRequest sends content related requests
// this function is used for getting, updating and deleting content
func (a *API) SendPageRequest(ep *url.URL, method string, c *Page) (*Page, error) {

	var body io.Reader
	if c != nil {
		js, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(js))
	}

	req, err := http.NewRequest(method, ep.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	res, err := a.Request(req)
	if err != nil {
		return nil, err
	}

	var content Page
	if len(res) != 0 {
		err = json.Unmarshal(res, &content)
		if err != nil {
			return nil, err
		}
	}
	return &content, nil
}

// SendPagesRequest sends content related requests
// this function is used for listing
func (a *API) SendPagesRequest(ep *url.URL, method string, c *Page) (*Pages, error) {

	var body io.Reader
	if c != nil {
		js, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(js))
	}

	req, err := http.NewRequest(method, ep.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	res, err := a.Request(req)
	if err != nil {
		return nil, err
	}

	var content Pages
	if len(res) != 0 {
		err = json.Unmarshal(res, &content)
		if err != nil {
			return nil, err
		}
	}
	return &content, nil
}
