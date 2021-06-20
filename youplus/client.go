package youplus

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"youfile/config"
)

var DefaultClient *Client

func InitClient() error {
	DefaultClient = &Client{
		baseUrl: config.Instance.YouPlusUrl,
		client:  resty.New(),
	}
	info, err := DefaultClient.GetInfo()
	if err != nil {
		return err
	}
	if !info.Success {
		return errors.New("get info not successful")
	}
	return nil
}

type Client struct {
	baseUrl string
	client  *resty.Client
}

type GetRealPathResponseBody struct {
	Path string `json:"path"`
}

func (c *Client) GetRealPath(target string, token string) (string, error) {
	var responseBody GetRealPathResponseBody
	_, err := c.client.R().
		SetQueryParam("target", target).
		SetHeader("Authorization", token).
		SetResult(&responseBody).
		Get(c.baseUrl + "/path/realpath")
	return responseBody.Path, err
}

type GetInfoResponseBody struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
}

func (c *Client) GetInfo() (*GetInfoResponseBody, error) {
	var responseBody GetInfoResponseBody
	_, err := c.client.R().
		SetResult(&responseBody).
		Get(c.baseUrl + "/info")
	return &responseBody, err
}

type ReadDirItem struct {
	RealPath string `json:"realPath"`
	Path     string `json:"path"`
	Type     string `json:"type"`
}

func (c *Client) ReadDir(target string, token string) ([]ReadDirItem, error) {
	var responseBody []ReadDirItem
	_, err := c.client.R().
		SetQueryParam("target", target).
		SetHeader("Authorization", token).
		SetResult(&responseBody).
		Get(c.baseUrl + "/path/readdir")
	return responseBody, err
}
