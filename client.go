package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
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

func handleBody(body any) (*bytes.Buffer, error) {
	var payloadReader *bytes.Buffer
	if body != nil {
		payloadReader = &bytes.Buffer{}
		if err := json.NewEncoder(payloadReader).Encode(body); err != nil {
			return nil, fmt.Errorf("error while json encoding body: %w", err)
		}
	}
	return payloadReader, nil
}

func (c *client) httpPost(api string, body any) (*http.Response, error) {
	return c.request("POST", path.Join(c.serverURL, api), body)
}

func (c *client) httpGet(api string, body any) (*http.Response, error) {
	return c.request("GET", path.Join(c.serverURL, api), body)
}

func (c *client) request(method, url string, body any) (*http.Response, error) {
	payload, err := handleBody(body)
	if err != nil {
		return nil, err
	}
	var httpReq *http.Request
	httpReq, err = http.NewRequest(method, url, payload)
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
	authReq, err := http.NewRequest("POST", path.Join(c.serverURL, "/api/auth/client/login"), body)
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
	task := &Task{}
	if err = json.NewDecoder(res.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("error while reading migration response: %w", err)
	}
	return task, nil
}

func (c *client) GetMigrationStatus(taskID string) (*Task, error) {
	res, err := c.httpPost(path.Join("/api/migrate", taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("error while getting migration status: %w", err)
	}
	task := &Task{}
	if err = json.NewDecoder(res.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("error while reading migration response: %w", err)
	}
	return task, nil
}
