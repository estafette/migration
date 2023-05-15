package migration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	contracts "github.com/estafette/estafette-ci-contracts"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	requestTimeout = 60 * time.Second
	// tokenTimeout based on https://github.com/estafette/estafette-ci-api/blob/main/pkg/api/middleware.go#L68
	tokenTimeout = 175 * time.Minute
	migrationAPI = "/api/migrations"
)

var (
	ErrAuthFailed = fmt.Errorf("authentication failed")
)

// Client for the estafette-ci-api migration API
type Client interface {
	Queue(request Request) (*Task, error)
	GetMigrationByID(taskID string) (*Task, error)
	RollbackMigration(taskID string) (*Changes, error)
	GetMigrations() ([]*Task, error)
	GetMigrationByFromRepo(source, owner, name string) (*Task, error)
	GetPipelineBuildStatus(source, owner, name, branch string) (string, error)
}

type bearerAuth struct {
	clientID     string
	clientSecret string
	expiresIn    time.Time
	token        string
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	httpClient
	bearerAuth
	serverURL string
}

type authResponse struct {
	Token string `json:"token"`
}

// NewClient returns a new migration API Client for estafette-ci-api
func NewClient(serverURL, clientID, clientSecret string) Client {
	return &client{
		httpClient: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       requestTimeout,
		},
		bearerAuth: bearerAuth{
			clientID:     clientID,
			clientSecret: clientSecret,
		},
		serverURL: strings.TrimSuffix(serverURL, "/"),
	}
}

// httpPost request for the given api endpoint with optional body
func (c *client) httpPost(api string, body any) (*http.Response, error) {
	return c.request("POST", _urlJoin(c.serverURL, api), body)
}

// httpGet request for the given api endpoint with optional body
func (c *client) httpGet(api string, body any) (*http.Response, error) {
	return c.request("GET", _urlJoin(c.serverURL, api), body)
}

// httpDelete request for the given api endpoint with optional body
func (c *client) httpDelete(api string, body any) (*http.Response, error) {
	return c.request("DELETE", _urlJoin(c.serverURL, api), body)
}

// request handles request body encoding if provided and authentication if token has expired
func (c *client) request(method, url string, body any) (*http.Response, error) {
	var httpReq *http.Request
	var err error
	if body != nil {
		var payload []byte
		if payload, err = json.Marshal(body); err != nil {
			return nil, fmt.Errorf("error while json encoding body: %w", err)
		}
		httpReq, err = http.NewRequest(method, url, bytes.NewBuffer(payload))
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

// authenticate with estafette-ci-api using the clientID and clientSecret
func (c *client) authenticate() error {
	log.Debug().Str("module", "github.com/estafette/migration").Msgf("authenticating with estafette-ci-api using clientID %s", c.clientID)
	body := strings.NewReader(fmt.Sprintf(`{"clientID": "%s", "clientSecret": "%s"}`, c.clientID, c.clientSecret))
	authReq, err := http.NewRequest("POST", _urlJoin(c.serverURL, "/api/auth/client/login"), body)
	if err != nil {
		return fmt.Errorf("error while creating authentication request: %w", err)
	}
	var res *http.Response
	if res, err = c.Do(authReq); err != nil {
		return fmt.Errorf("error while authenticatiing: %w", err)
	}
	var data []byte
	if data, err = _successful(res); err != nil {
		return errors.Join(ErrAuthFailed, err)
	}
	var authResp authResponse
	if err = json.Unmarshal(data, &authResp); err != nil {
		return fmt.Errorf("error while reading auth response: %w", err)
	}
	c.token = authResp.Token
	c.expiresIn = time.Now().Add(tokenTimeout)
	return nil
}

// Queue task in estafette. If the ID of the task is not provided,
// it will be generated in Estafette server else existing task is updated
func (c *client) Queue(request Request) (*Task, error) {
	if request.CallbackURL != nil && *request.CallbackURL == "" {
		request.CallbackURL = nil
	}
	res, err := c.httpPost(migrationAPI, request)
	if err != nil {
		return nil, fmt.Errorf("queue api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return nil, fmt.Errorf("queue api: %w", err)
	}
	task := &Task{}
	if err = json.Unmarshal(body, task); err != nil {
		return nil, fmt.Errorf("queue api: error while unmarshalling response: %w", err)
	}
	return task, nil
}

// GetMigrationByID of migration task using task ID
func (c *client) GetMigrationByID(taskID string) (*Task, error) {
	res, err := c.httpGet(_urlJoin(migrationAPI, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("getMigrationByID api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return nil, fmt.Errorf("getMigrationByID api: %w", err)
	}
	task := &Task{}
	if err = json.Unmarshal(body, task); err != nil {
		return nil, fmt.Errorf("getStatus api: error while unmarshalling response: %w", err)
	}
	return task, nil
}

// RollbackMigration task in estafette.
func (c *client) RollbackMigration(taskID string) (*Changes, error) {
	res, err := c.httpDelete(_urlJoin(migrationAPI, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("rollbackMigration api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return nil, fmt.Errorf("rollbackMigration api: %w", err)
	}
	changes := &Changes{}
	if err = json.Unmarshal(body, changes); err != nil {
		return nil, fmt.Errorf("rollbackMigration api: error while unmarshalling response: %w", err)
	}
	return changes, nil
}

// GetMigrationByFromRepo of migration task using task ID
func (c *client) GetMigrationByFromRepo(source, owner, name string) (*Task, error) {
	res, err := c.httpGet(_urlJoin(migrationAPI, "from", source, owner, name), nil)
	if err != nil {
		return nil, fmt.Errorf("getMigrationByFromRepo api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return nil, fmt.Errorf("getMigrationByFromRepo api: %w", err)
	}
	task := &Task{}
	if err = json.Unmarshal(body, task); err != nil {
		return nil, fmt.Errorf("getStatus api: error while unmarshalling response: %w", err)
	}
	return task, nil
}

func (c *client) GetMigrations() ([]*Task, error) {
	res, err := c.httpGet(migrationAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("getMigrations api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return nil, fmt.Errorf("getMigrations api: %w", err)
	}
	tasks := make([]*Task, 0)
	if err = json.Unmarshal(body, &tasks); err != nil {
		return nil, fmt.Errorf("getMigrations api: error while unmarshalling response: %w", err)
	}
	return tasks, nil
}

// GetPipelineBuildStatus of pipeline using source, owner and name
func (c *client) GetPipelineBuildStatus(source, owner, name, branch string) (string, error) {
	res, err := c.httpGet(_urlJoin(source, owner, name, "builds"), nil)
	if err != nil {
		return "", fmt.Errorf("getPipelineStatus api: error while executing request: %w", err)
	}
	var body []byte
	body, err = _successful(res)
	if err != nil {
		return "", fmt.Errorf("getPipelineStatus api: %w", err)
	}

	var builds []*contracts.Build
	if err = json.Unmarshal(body, &builds); err != nil {
		return "", fmt.Errorf("getPipelineStatus api: error while unmarshalling response: %w", err)
	}

	// Get the latest build for the branch
	for _, build := range builds {
		if build.RepoBranch == branch {
			return string(build.BuildStatus), nil
		}
	}
	return "", nil
}
