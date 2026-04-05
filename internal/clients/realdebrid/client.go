package realdebrid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://api.real-debrid.com/rest/1.0"

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *Client) UnrestrictLink(link string, password string, remote *int) (*UnrestrictResponse, error) {
	form := url.Values{}
	form.Set("link", link)
	if password != "" {
		form.Set("password", password)
	}
	if remote != nil {
		form.Set("remote", fmt.Sprintf("%d", *remote))
	}

	req, err := http.NewRequest("POST", baseURL+"/unrestrict/link", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB cap

	if resp.StatusCode != http.StatusOK {
		var rdErr ErrorResponse
		if json.Unmarshal(body, &rdErr) == nil && rdErr.Message != "" {
			return nil, rdErr
		}
		return nil, fmt.Errorf("realdebrid: http %d: %s", resp.StatusCode, string(body))
	}

	var out UnrestrictResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Download == "" {
		return nil, errors.New("realdebrid: missing download url")
	}

	return &out, nil
}

func (c *Client) DownloadFile(downloadURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: http %d", resp.StatusCode)
	}

	return resp, nil
}
