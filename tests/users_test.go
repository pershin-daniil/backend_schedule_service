package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/notifier"
	"github.com/pershin-daniil/TimeSlots/pkg/pgstore"
	"github.com/pershin-daniil/TimeSlots/pkg/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
)

const (
	testURL = "http://localhost:8080"
	address = ":8080"
	version = "test"
	pgDSN   = "postgres://postgres:secret@localhost:6431/timeslots?sslmode=disable"
)

var user = models.User{
	LastName:  "Ivanov",
	FirstName: "Ivan",
}

type IntegrationTestSuite struct {
	suite.Suite
	log      *logrus.Logger
	store    service.Store
	notifier service.Notifier
	app      rest.App
	handler  *rest.Server
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logger.NewLogger()
	var err error
	ctx := context.Background()
	s.store, err = pgstore.NewStore(ctx, s.log, pgDSN)
	s.notifier = notifier.NewDummyNotifier(s.log)
	s.app = service.NewScheduleService(s.log, s.store, s.notifier)
	s.Require().NoError(err)

	s.handler = rest.NewServer(s.log, s.app, address, version)
	go func() {
		_ = s.handler.Run()
	}()
	time.Sleep(100 * time.Millisecond)
	err = s.store.TruncateTable(ctx, "users")
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) createUser(ctx context.Context, user models.User) int {
	s.T().Helper()
	result := models.User{}
	resp := s.sendRequest(ctx, http.MethodPost, "/api/v1/users", user, &result)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	return result.ID
}

func (s *IntegrationTestSuite) updateUser(ctx context.Context, data models.User, id int) models.User { //nolint:unused
	s.T().Helper()
	result := models.User{}
	resp := s.sendRequest(ctx, http.MethodPatch, "/api/v1/users/"+strconv.Itoa(id), data, &result)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	return result
}

func (s *IntegrationTestSuite) deleteUser(ctx context.Context, id int) models.User {
	s.T().Helper()
	result := models.User{}
	resp := s.sendRequest(ctx, http.MethodDelete, "/api/v1/users/"+strconv.Itoa(id), nil, &result)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	return result
}

func (s *IntegrationTestSuite) TestCreateUser() {
	ctx := context.Background()
	s.Run("create user", func() {
		reqBody, err := json.Marshal(user)
		s.Require().NoError(err)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+"/api/v1/users", bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusCreated, resp.StatusCode)
		var respUser models.User
		err = json.NewDecoder(resp.Body).Decode(&respUser)
		s.Require().NoError(err)
		s.Require().Equal(user.LastName, respUser.LastName)
		s.Require().Equal(user.FirstName, respUser.FirstName)
	})
}

func (s *IntegrationTestSuite) TestGetUser() {
	ctx := context.Background()
	id := s.createUser(ctx, user)

	s.Run("get user", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/users/"+strconv.Itoa(id), nil)
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respUser models.User
		err = json.NewDecoder(resp.Body).Decode(&respUser)
		s.Require().NoError(err)
		s.Require().Equal(id, respUser.ID)
		s.Require().Equal(user.LastName, respUser.LastName)
		s.Require().Equal(user.FirstName, respUser.FirstName)
	})

	s.Run("get user not found", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/users/0", nil)
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
	})
}

func (s *IntegrationTestSuite) TestUpdateUser() {
	ctx := context.Background()
	data := models.User{
		LastName:  "Updated",
		FirstName: "Booop!",
	}
	id := s.createUser(ctx, user)
	reqBody, err := json.Marshal(data)
	s.Require().NoError(err)

	s.Run("update user", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/users/"+strconv.Itoa(id), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respUser models.User
		err = json.NewDecoder(resp.Body).Decode(&respUser)
		s.Require().NoError(err)
		s.Require().Equal(id, respUser.ID)
		s.Require().Equal(data.LastName, respUser.LastName)
		s.Require().Equal(data.FirstName, respUser.FirstName)
	})

	s.Run("update user not found", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/users/0", bytes.NewReader(reqBody))
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
	})
}

func (s *IntegrationTestSuite) TestDeleteUser() {
	ctx := context.Background()
	id := s.createUser(ctx, user)

	s.Run("delete user", func() {
		reqBody, err := json.Marshal(user)
		s.Require().NoError(err)
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/users/"+strconv.Itoa(id), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respUser models.User
		err = json.NewDecoder(resp.Body).Decode(&respUser)
		s.Require().NoError(err)
		s.Require().Equal(id, respUser.ID)
		s.Require().Equal(user.LastName, respUser.LastName)
	})

	s.Run("delete user not found", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/users/0", nil)
		s.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
	})
}

func (s *IntegrationTestSuite) sendRequest(ctx context.Context, method, url string, body interface{}, dest interface{}) *http.Response {
	s.T().Helper()
	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(ctx, method, testURL+url, bytes.NewReader(reqBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()
	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}
	return resp
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
