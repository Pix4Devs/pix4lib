package proxyscraper

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/corpix/uarand"
)

type (
	Client struct {
		*http.Client
	}
)

const (
	API = "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=all&timeout=10000&country=all&ssl=all&anonymity=all"
)

func NewClient(timeout time.Duration) *Client {
	return &Client{
		&http.Client{
			Transport: &http.Transport{
				ForceAttemptHTTP2: true,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: timeout,
		},
	}
}

func(c *Client) Execute() ([]string, error) {
	var body io.ReadCloser

	req, err := http.NewRequest("GET", API, body); if err != nil {
		return nil, err
	}

	{
		req.Header.Add("content-type", "text/plain")
		req.Header.Add("user-agent", uarand.GetRandom())
	}
	
	resp, err := c.Client.Do(req); if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("[ERROR] %s responded with %s", API, resp.Status))
	}

	data, err := io.ReadAll(resp.Body); if err != nil {
		return nil, err
	}
	
	return strings.Split(strings.Trim(strings.TrimSpace(string(data)), "\r"),"\n"), nil
}