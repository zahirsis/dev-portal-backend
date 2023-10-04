package confluenceapiv2

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func NewAPI(location string, username string, token string) (*API, error) {
	if len(location) == 0 {
		return nil, errors.New("url empty")
	}

	u, err := url.ParseRequestURI(location + "/wiki/api/v2")

	if err != nil {
		return nil, err
	}

	a := new(API)
	a.endPoint = u
	a.token = token
	a.username = username

	// #nosec G402
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	a.Client = &http.Client{Transport: tr}

	return a, nil
}

// VerifyTLS to enable disable certificate checks
func (a *API) VerifyTLS(set bool) {
	// #nosec G402
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !set},
	}
	a.Client = &http.Client{Transport: tr}
}

// DebugFlag is the global debugging variable
var DebugFlag = false

// SetDebug enables debug output
func SetDebug(state bool) {
	DebugFlag = state
}

// Debug outputs debug messages
func Debug(msg interface{}) {
	if DebugFlag {
		fmt.Printf("%+v\n", msg)
	}
}
