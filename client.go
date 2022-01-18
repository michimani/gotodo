package gotodo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github/michimani/gotodo/util"

	"github.com/google/uuid"
)

const (
	APITokenEnvKey string = "GOTODO_API_TOKEN"
)

type NewClientInput struct {
	HTTPClient *http.Client
	APIToken   string
}

type Client struct {
	HTTPClient *http.Client
	apiToken   string
}

var defaultHTTPClient = &http.Client{
	Timeout: time.Duration(30) * time.Second,
}

func NewClient(in *NewClientInput) (*Client, error) {
	if in == nil {
		return nil, fmt.Errorf("NewClientInput is nil.")
	}

	c := Client{}
	if in.HTTPClient != nil {
		c.HTTPClient = in.HTTPClient
	} else {
		c.HTTPClient = defaultHTTPClient
	}

	c.SetAPIToken(in.APIToken)

	return &c, nil
}

func (c *Client) APIToken() string {
	return c.apiToken
}

func (c *Client) SetAPIToken(token string) {
	c.apiToken = token
}

func (c *Client) IsReady() bool {
	if c == nil {
		return false
	}

	if c.APIToken() == "" {
		return false
	}

	return true
}

func (c *Client) CallAPI(ctx context.Context, endpoint, method string, p util.Parameters, r util.Response) error {
	req, err := c.prepare(ctx, endpoint, method, p)
	if err != nil {
		return err
	}

	err = c.exec(req, r)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) exec(req *http.Request, r util.Response) error {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return err
	}

	return nil
}

func (c *Client) prepare(ctx context.Context, endpointBase, method string, p util.Parameters) (*http.Request, error) {
	if p == nil {
		return nil, fmt.Errorf("parameters is nil")
	}

	if !c.IsReady() {
		return nil, fmt.Errorf("client is not ready")
	}

	endpoint := p.ResolveEndpoint(endpointBase)
	req, err := newRequest(ctx, endpoint, method, p)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken()))

	return req, nil
}

func newRequest(ctx context.Context, endpoint, method string, p util.Parameters) (*http.Request, error) {
	body, err := p.Body()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	if method == http.MethodPost {
		reqUUID, err := generateUUID()
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Request-Id", reqUUID)
	}

	return req, nil
}

func generateUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	uu := u.String()
	return uu, nil
}
