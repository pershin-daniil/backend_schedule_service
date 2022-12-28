package tests

import (
	"bytes"
	"context"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/pershin-daniil/TimeSlots/pkg/store"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
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
	require.NoError(s.T(), err)
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
		LastName:    "TestLN",
		FirstName:   "TestFN",
		PhoneNumber: 1234567890,
	}
	reqBody, err := json.Marshal(user)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+"/api/v1/createUser", bytes.NewReader(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(s.T(), err)
	}()
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	var respUser models.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	require.NoError(s.T(), err)
	require.Equal(s.T(), user.LastName, respUser.LastName)
	require.Equal(s.T(), user.FirstName, respUser.FirstName)
	require.Equal(s.T(), user.PhoneNumber, respUser.PhoneNumber)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
