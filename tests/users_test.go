package tests

import (
	"bytes"
	"context"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/store"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
)

const (
	testURL = "http://localhost:8080"
	address = ":8080"
	version = "test"
	pgDSN   = "postgres://postgres:secret@localhost:6431/timeslots?sslmode=disable"
)

type IntegrationTestSuite struct {
	suite.Suite
	log     *logrus.Logger
	app     *store.Store
	handler *rest.Handler
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logger.NewLogger()
	var err error
	s.app, err = store.NewStore(s.log, pgDSN)
	s.Require().NoError(err)
	s.handler = rest.NewHandler(s.log, s.app, address, version)
	go func() {
		_ = s.handler.Run()
	}()
	time.Sleep(100 * time.Millisecond)
}

func (s *IntegrationTestSuite) SetupTest() {
	ctx := context.Background()
	err := s.app.TruncateTable(ctx, "users")
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) createUser(ctx context.Context, user models.User) int {
	s.T().Helper()
	result := models.User{}
	resp := s.sendRequest(ctx, http.MethodPost, "/api/v1/users", user, &result)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	return result.ID
}

func (s *IntegrationTestSuite) TestCreateUser() {
	ctx := context.Background()
	user := models.User{
		LastName:  "TestLN",
		FirstName: "TestFN",
	}
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
}

func (s *IntegrationTestSuite) TestGetUser() {
	ctx := context.Background()
	user := models.User{
		LastName:  "TestLN",
		FirstName: "TestFN",
	}
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
