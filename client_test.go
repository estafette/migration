package migration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	rArgs := c.Called(req)
	return rArgs.Get(0).(*http.Response), rArgs.Error(1)
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:80/", "clientID", "clientSecret")
	shouldBe := assert.New(t)
	shouldBe.NotNil(c)
	shouldBe.Equal("http://localhost:80", c.(*client).serverURL)
}

func TestClient_Queue_Success(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"test-1","fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new","status":"queued","lastStep":"waiting"}`))}, nil).
		Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"test-1","fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new","status":"queued","lastStep":"waiting"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	req := Request{FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"}
	_, err := c.Queue(req)
	shouldBe.Nil(err)
	task, err := c.Queue(req)
	if shouldBe.Nil(err) {
		shouldBe.Equal(&Task{
			Request:  Request{ID: "test-1", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
			Status:   StatusQueued,
			LastStep: StepWaiting,
		}, task)
	}
	if mockedClient.AssertExpectations(t) {
		migrationReq := mockedClient.Calls[1].Arguments[0].(*http.Request)
		shouldBe.NotNil(migrationReq)
		shouldBe.Equal("POST", migrationReq.Method)
		shouldBe.Equal("http://localhost:80/api/migrations", migrationReq.URL.String())
		data, err := io.ReadAll(migrationReq.Body)
		shouldBe.Nil(err)
		shouldBe.Equal(`{"fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new"}`, string(data))
	}
}

func TestClient_Queue_Failure(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.On("Do", mock.Anything).
		Return(&http.Response{Status: "404 Not Found", StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"code":404,"message":"Pipeline not found"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	req := Request{FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"}
	task, err := c.Queue(req)
	shouldBe.Nil(task)
	shouldBe.Equal(fmt.Errorf("queue api: %w", fmt.Errorf(`responded with status: 404 Not Found, body: {"code":404,"message":"Pipeline not found"}`)), err)
}

func TestClient_GetMigrationByID_Success(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"test-123","fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new","status":"in_progress","lastStep":"releases_done"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	task, err := c.GetMigrationByID("test-123")
	if shouldBe.Nil(err) {
		shouldBe.Equal(&Task{
			Request:  Request{ID: "test-123", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
			Status:   StatusInProgress,
			LastStep: StepReleasesDone,
		}, task)
	}
	if mockedClient.AssertExpectations(t) {
		migrationReq := mockedClient.Calls[1].Arguments[0].(*http.Request)
		shouldBe.NotNil(migrationReq)
		shouldBe.Equal("GET", migrationReq.Method)
		shouldBe.Equal("http://localhost:80/api/migrations/test-123", migrationReq.URL.String())
		shouldBe.Nil(migrationReq.Body)
	}
}

func TestClient_GetMigrationByID_Failure(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.On("Do", mock.Anything).
		Return(&http.Response{Status: "404 Not Found", StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"code":404,"message":"migration task not found"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	task, err := c.GetMigrationByID("test-123")
	shouldBe.Nil(task)
	shouldBe.Equal(fmt.Errorf("getMigrationByID api: %w", fmt.Errorf(`responded with status: 404 Not Found, body: {"code":404,"message":"migration task not found"}`)), err)
}

func TestClient_GetMigrationByFromRepo_Success(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"test-123","fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new","status":"in_progress","lastStep":"releases_done"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	task, err := c.GetMigrationByFromRepo("github.com", "estafette", "migration")
	if shouldBe.Nil(err) {
		shouldBe.Equal(&Task{
			Request:  Request{ID: "test-123", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
			Status:   StatusInProgress,
			LastStep: StepReleasesDone,
		}, task)
	}
	if mockedClient.AssertExpectations(t) {
		migrationReq := mockedClient.Calls[1].Arguments[0].(*http.Request)
		shouldBe.NotNil(migrationReq)
		shouldBe.Equal("GET", migrationReq.Method)
		shouldBe.Equal("http://localhost:80/api/migrations/from/github.com/estafette/migration", migrationReq.URL.String())
		shouldBe.Nil(migrationReq.Body)
	}
}

func TestClient_GetMigrationByFromRepo_Failure(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.On("Do", mock.Anything).
		Return(&http.Response{Status: "404 Not Found", StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"code":404,"message":"migration task not found"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	task, err := c.GetMigrationByFromRepo("github.com", "estafette", "migration")
	shouldBe.Nil(task)
	shouldBe.Equal(fmt.Errorf("getMigrationByFromRepo api: %w", fmt.Errorf(`responded with status: 404 Not Found, body: {"code":404,"message":"migration task not found"}`)), err)
}

func TestClient_Rollback_Success(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"releases": 10,"releaseLogs": 11,"builds": 12,"buildLogs": 13,"buildVersions": 14}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	changes, err := c.RollbackMigration("test-123")
	if shouldBe.Nil(err) {
		shouldBe.Equal(&Changes{
			Releases:      10,
			ReleaseLogs:   11,
			Builds:        12,
			BuildLogs:     13,
			BuildVersions: 14,
		}, changes)
	}
	if mockedClient.AssertExpectations(t) {
		migrationReq := mockedClient.Calls[1].Arguments[0].(*http.Request)
		shouldBe.NotNil(migrationReq)
		shouldBe.Equal("DELETE", migrationReq.Method)
		shouldBe.Equal("http://localhost:80/api/migrations/test-123", migrationReq.URL.String())
		shouldBe.Nil(migrationReq.Body)
	}
}

func TestClient_Rollback_Failure(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.On("Do", mock.Anything).
		Return(&http.Response{Status: "404 Not Found", StatusCode: 404, Body: io.NopCloser(strings.NewReader(`{"code":404,"message":"migration task not found"}`))}, nil).
		Once()
	shouldBe := assert.New(t)
	changes, err := c.RollbackMigration("test-123")
	shouldBe.Nil(changes)
	shouldBe.Equal(fmt.Errorf("rollbackMigration api: %w", fmt.Errorf(`responded with status: 404 Not Found, body: {"code":404,"message":"migration task not found"}`)), err)
}

func TestClient_GetMigrations_Success(t *testing.T) {
	mockedClient := &mockClient{}
	c := &client{
		httpClient: mockedClient,
		bearerAuth: bearerAuth{
			clientID:     "test-clientID",
			clientSecret: "test-clientSecret",
		},
		serverURL: "http://localhost:80",
	}
	mockAuth(mockedClient).Once()
	mockedClient.
		On("Do", mock.Anything).
		Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[
			{"id":"test-123","fromSource":"github.com","fromOwner":"estafette","fromName":"migration","toSource":"github.com","toOwner":"estafette_new","toName":"migration_new","status":"in_progress","lastStep":"releases_done"},
			{"id":"test-345","fromSource":"github.com","fromOwner":"estafette","fromName":"m_grati@n","toSource":"github.com","toOwner":"estafette_new","toName":"m_grati@n_new","status":"in_progress","lastStep":"releases_done"}
]`))}, nil).
		Once()
	shouldBe := assert.New(t)
	tasks, err := c.GetMigrations()
	if shouldBe.Nil(err) {
		shouldBe.Equal([]*Task{
			{
				Request:  Request{ID: "test-123", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
				Status:   StatusInProgress,
				LastStep: StepReleasesDone,
			}, {
				Request:  Request{ID: "test-345", FromSource: "github.com", FromOwner: "estafette", FromName: "m_grati@n", ToSource: "github.com", ToOwner: "estafette_new", ToName: "m_grati@n_new"},
				Status:   StatusInProgress,
				LastStep: StepReleasesDone,
			},
		}, tasks)
	}
	if mockedClient.AssertExpectations(t) {
		migrationReq := mockedClient.Calls[1].Arguments[0].(*http.Request)
		shouldBe.NotNil(migrationReq)
		shouldBe.Equal("GET", migrationReq.Method)
		shouldBe.Equal("http://localhost:80/api/migrations", migrationReq.URL.String())
		shouldBe.Nil(migrationReq.Body)
	}
}

func mockAuth(mockedClient *mockClient) *mock.Call {
	return mockedClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://localhost:80/api/auth/client/login"
	})).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"token":"test-token"}`))}, nil)
}
