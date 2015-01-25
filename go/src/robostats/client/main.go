// Package client is a Go consumer for robotstats.
package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	endpointPrefix = "http://api.dev.robostats.io"
)

const (
	loginEndpoint               = endpointPrefix + "/user/login"
	deviceClassIndexEndpoint    = endpointPrefix + "/device_classes"
	deviceInstanceIndexEndpoint = endpointPrefix + "/device_instances"
	deviceSessionIndexEndpoint  = endpointPrefix + "/device_sessions"
)

const (
	MethodPost = "POST"
)

type Authorization struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type Client struct {
	Email    string
	Password string
	Auth     Authorization
}

func (a *Authorization) authHeader() string {
	return strings.Title(a.TokenType) + " " + a.AccessToken
}

func decodeBody(res *http.Response, dest interface{}) error {
	var buf []byte
	var err error

	if buf, err = ioutil.ReadAll(res.Body); err != nil {
		return err
	}

	if err := json.Unmarshal(buf, &dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) Login() error {
	var res *http.Response
	var err error

	params := url.Values{
		"email":    {c.Email},
		"password": {c.Password},
	}

	if res, err = http.PostForm(loginEndpoint, params); err != nil {
		return err
	}

	if err = decodeBody(res, &c.Auth); err != nil {
		return err
	}

	return nil
}

func (c *Client) signRequest(req *http.Request) {
	req.Header.Add("Authorization", c.Auth.authHeader())
}

func (c *Client) signedGet(endpoint string, values url.Values, dest interface{}) error {
	var req *http.Request
	var res *http.Response
	var err error

	cli := http.Client{}

	var uri *url.URL

	if uri, err = url.Parse(endpoint); err != nil {
		return err
	}

	uri.RawQuery = values.Encode()

	if req, err = http.NewRequest("GET", uri.String(), nil); err != nil {
		return err
	}

	c.signRequest(req)

	if res, err = cli.Do(req); err != nil {
		return err
	}

	if err = decodeBody(res, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetClasses() ([]Class, error) {
	var data classesEnvelope

	if err := c.signedGet(deviceClassIndexEndpoint, nil, &data); err != nil {
		return nil, err
	}

	return data.Classes, nil
}

func (c *Client) GetInstancesByClassID(classID string) ([]Instance, error) {
	var data instancesEnvelope

	if err := c.signedGet(deviceInstanceIndexEndpoint, url.Values{"device_class_id": {classID}}, &data); err != nil {
		return nil, err
	}

	return data.Instances, nil
}

func (c *Client) GetSessionsByInstanceID(instanceID string) ([]Session, error) {
	var data sessionsEnvelope

	if err := c.signedGet(deviceSessionIndexEndpoint, url.Values{"device_instance_id": {instanceID}}, &data); err != nil {
		return nil, err
	}

	return data.Sessions, nil
}
