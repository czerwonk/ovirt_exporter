package api

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ApiClient struct {
	Url      string
	Username string
	Password string
	client   *http.Client
}

func NewClient(url, username, password string, insecureCert bool) *ApiClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}

	return &ApiClient{Url: url, Username: username, Password: password, client: c}
}

func (c *ApiClient) GetAndParse(path string, v interface{}) error {
	b, err := c.Get(path)

	if err != nil {
		return err
	}

	err = xml.Unmarshal(b, v)
	return err
}

func (c *ApiClient) Get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", strings.Trim(c.Url, "/")+strings.Trim(path, "/"), nil)

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

func StatiscticsPath(ressourceType string, id string) string {
	return fmt.Sprintf("/%s/%s/statistics", ressourceType, id)
}
