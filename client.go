package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// tokenTimeout based on https://github.com/estafette/estafette-ci-api/blob/main/pkg/api/middleware.go#L68
	tokenTimeout = 175 * time.Minute
)

type Client interface {
	QueueMigration(request TaskRequest) (*Task, error)
	GetMigrationStatus(taskID string) (*Task, error)
}

type bearerAuth struct {
	clientID     string
	clientSecret string
	expiresIn    time.Time
	token        string
}

type client struct {
	*http.Client
	bearerAuth
	serverURL string
}

type authResponse struct {
	Token string `json:"token"`
}

func NewClient(serverURL, clientID, clientSecret string) Client {
	return &client{
		Client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       60 * time.Second,
		},
		bearerAuth: bearerAuth{
			clientID:     clientID,
			clientSecret: clientSecret,
		},
		serverURL: strings.TrimSuffix(serverURL, "/"),
	}
}

func urlJoin(URL string, path ...string) string {
	if u, err := url.JoinPath(URL, path...); err != nil {
		panic(err)
	} else {
		return u
	}
}

func (c *client) httpPost(api string, body any) (*http.Response, error) {
	return c.request("POST", urlJoin(c.serverURL, api), body)
}

func (c *client) httpGet(api string, body any) (*http.Response, error) {
	return c.request("GET", urlJoin(c.serverURL, api), body)
}

func (c *client) request(method, url string, body any) (*http.Response, error) {
	var httpReq *http.Request
	var err error
	if body != nil {
		payload := &bytes.Buffer{}
		if err := json.NewEncoder(payload).Encode(body); err != nil {
			return nil, fmt.Errorf("error while json encoding body: %w", err)
		}
		httpReq, err = http.NewRequest(method, url, payload)
	} else {
		httpReq, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error while creating http request [%s]%s %v: %w", method, url, body, err)
	}
	if time.Now().After(c.expiresIn) {
		if err = c.authenticate(); err != nil {
			return nil, err
		}
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Add("Content-Type", "application/json")
	var res *http.Response
	res, err = c.Do(httpReq)
	if err != nil {
		return res, fmt.Errorf("error while executing http request [%s]%s %v: %w", method, url, body, err)
	}
	return res, nil
}

func (c *client) authenticate() error {
	body := strings.NewReader(fmt.Sprintf(`{"clientID": "%s", "clientSecret": "%s"}`, c.clientID, c.clientSecret))
	authReq, err := http.NewRequest("POST", urlJoin(c.serverURL, "/api/auth/client/login"), body)
	if err != nil {
		return fmt.Errorf("error while creating authentication request: %w", err)
	}
	var resp *http.Response
	resp, err = c.Do(authReq)
	if err != nil {
		return fmt.Errorf("error while authenticatiing: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Println("Error while closing the response body:", err)
		}
	}(resp.Body)
	var authResp authResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return fmt.Errorf("error while reading auth response: %w", err)
	}
	c.token = authResp.Token
	c.expiresIn = time.Now().Add(tokenTimeout)
	return nil
}

func (c *client) QueueMigration(request TaskRequest) (*Task, error) {
	if request.CallbackURL != nil && *request.CallbackURL == "" {
		request.CallbackURL = nil
	}
	res, err := c.httpPost("/api/migrate", request)
	if err != nil {
		return nil, fmt.Errorf("error while queuing migration: %w", err)
	}
	var body []byte
	body, err = isResponseOK(res)
	if err != nil {
		if body != nil {
			log.Printf("error response from /api/migrate: %s", string(body))
		}
		return nil, err
	}
	task := &Task{}
	if err = json.Unmarshal(body, task); err != nil {
		return nil, fmt.Errorf("error while reading migration response: %w", err)
	}
	return task, nil
}

func (c *client) GetMigrationStatus(taskID string) (*Task, error) {
	res, err := c.httpGet(urlJoin("/api/migrate", taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("error while getting migration status: %w", err)
	}
	var body []byte
	body, err = isResponseOK(res)
	if err != nil {
		if body != nil {
			log.Printf("error response from /api/migrate: %s", string(body))
		}
		return nil, err
	}
	task := &Task{}
	if err = json.Unmarshal(body, task); err != nil {
		return nil, fmt.Errorf("error while reading migration response: %w", err)
	}
	return task, nil
}

func isResponseOK(res *http.Response) ([]byte, error) {
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Println("Error while closing the response body:", err)
		}
	}(res.Body)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading resposne body: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return bodyBytes, fmt.Errorf("response with status: %s", res.Status)
	}
	return bodyBytes, nil
}
