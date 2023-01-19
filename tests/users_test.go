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

func (s *IntegrationTestSuite) TestCreateUser() {
	ctx := context.Background()
	user := models.User{
		ID:        3,
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
