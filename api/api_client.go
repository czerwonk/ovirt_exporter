package api

import (
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"

	"errors"
)

// ApiClient encapsulates communication with the oVirt REST API
type ApiClient struct {
	Url      string
	Username string
	Password string
	client   *http.Client
	logger   Logger
}

// NewClient returns a new client
func NewClient(url, username, password string, insecureCert bool, logger Logger) *ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}

	return &ApiClient{Url: url, Username: username, Password: password, client: c, logger: logger}
}

// GetAndParse retrieves XML data from the API and unmarshals it
func (c *ApiClient) GetAndParse(path string, v interface{}) error {
	b, err := c.Get(path)

	if err != nil {
		return err
	}

	err = xml.Unmarshal(b, v)
	return err
}

// Get retrieves XML data from the API and returns it
func (c *ApiClient) Get(path string) ([]byte, error) {
	uri := strings.Trim(c.Url, "/") + "/" + strings.Trim(path, "/")
	c.logger.Debug("GET ", uri)
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
