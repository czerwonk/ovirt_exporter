package api

import (
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"errors"
	"fmt"
	"io"
)

// Client encapsulates communication with the oVirt REST API
type Client struct {
	URL      string
	Username string
	Password string
	Logger   Logger
	Debug    bool
	cookie   string
	client   *http.Client
}

// NewClient returns a new client
func NewClient(url, username, password string, insecureCert bool) (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureCert},
	}
	c := &http.Client{Transport: tr}
	logger := &DefaultLogger{}

	client := &Client{URL: url, Username: username, Password: password, client: c, Logger: logger}
	err := client.Auth()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Auth establishes a SSO session with oVirt API
func (c *Client) Auth() error {
	req, err := http.NewRequest("HEAD", c.URL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Prefer", "persistent-auth")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	c.cookie = strings.Split(resp.Header.Get("Set-Cookie"), ";")[0]
	return nil
}

// GetAndParse retrieves XML data from the API and unmarshals it
func (c *Client) GetAndParse(path string, v interface{}) error {
	return c.SendAndParse(path, "GET", v, nil)
}

// Get retrieves XML data from the API and returns it
func (c *Client) Get(path string) ([]byte, error) {
	return c.SendRequest(path, "GET", nil)
}

// Close terminates the SSO session with the API
func (c *Client) Close() {
	req, err := http.NewRequest("HEAD", c.URL, nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.Username, c.Password)
	c.client.Do(req)
}

// SendAndParse sends a request to the API and unmarshalls the response
func (c *Client) SendAndParse(path, method string, res interface{}, body io.Reader) error {
	b, err := c.SendRequest(path, method, body)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(b, res)
	return err
}

// SendRequest sends a request to the API
func (c *Client) SendRequest(path, method string, body io.Reader) ([]byte, error) {
	uri := strings.Trim(c.URL, "/") + "/" + strings.Trim(path, "/")
	c.Logger.Debug(method, uri)

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/xml")
	req.Header.Set("Prefer", "persistent-auth")
	req.Header.Set("Cookie", c.cookie)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println(resp.Status)
	if c.Debug {
		c.Logger.Debug(string(b))
	}

	return b, err
}
