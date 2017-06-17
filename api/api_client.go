package api

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/prometheus/common/log"
)

// ApiClient encapsulates communication with the oVirt REST API
type ApiClient struct {
	Url      string
	Username string
	Password string
	client   *http.Client
}

// NewClient returns a new client
func NewClient(url, username, password string, insecureCert bool) *ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}

	return &ApiClient{Url: url, Username: username, Password: password, client: c}
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
	log.Debug("GET ", uri)
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// StatiscticsPath builds an statistics path for a given ressource
func StatiscticsPath(ressourceType string, id string) string {
	return fmt.Sprintf("/%s/%s/statistics", ressourceType, id)
}
