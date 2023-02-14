package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pershin-daniil/TimeSlots/internal/rest"
	"github.com/pershin-daniil/TimeSlots/pkg/logger"
	"github.com/pershin-daniil/TimeSlots/pkg/notifier"
	"github.com/pershin-daniil/TimeSlots/pkg/pgstore"
	"github.com/pershin-daniil/TimeSlots/pkg/service"
	migrate "github.com/rubenv/sql-migrate"
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

var user models.UserRequest

var meeting models.MeetingRequest

type errResp struct {
	Error string `json:"error"`
}

type IntegrationTestSuite struct {
	suite.Suite
	log      *logrus.Logger
	store    *pgstore.Store
	notifier service.Notifier
	app      rest.App
	handler  *rest.Server
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logger.NewLogger()
	var err error

	var (
		LastName     = "Ivanov"
		FirstName    = "Ivan"
		userPhone    = "+7 999 999 99 99"
		userEmail    = "example@mail.ru"
		userPassword = "jopa"
		userRole     = "client"
	)

	user = models.UserRequest{
		LastName:  &LastName,
		FirstName: &FirstName,
		Phone:     &userPhone,
		Email:     &userEmail,
		Password:  &userPassword,
		Role:      &userRole,
	}

	Manager := 0
	StartTime := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	EndTime := time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)
	Client := 0

	meeting = models.MeetingRequest{
		Manager:   &Manager,
		StartTime: &StartTime,
		EndTime:   &EndTime,
		Client:    &Client,
	}

	ctx := context.Background()
	s.store, err = pgstore.NewStore(ctx, s.log, pgDSN)
	s.Require().NoError(err)
	err = s.store.Migrate(migrate.Up)
	s.Require().NoError(err)
	s.notifier = notifier.NewDummyNotifier(s.log)
	s.app = service.NewScheduleService(s.log, s.store, s.notifier)
	s.Require().NoError(err)

	s.handler = rest.NewServer(s.log, s.app, address, version)
	go func() {
		_ = s.handler.Run(ctx)
	}()
	time.Sleep(100 * time.Millisecond)
	err = s.store.ResetTables(ctx, []string{"meetings", "users", "users_history", "meetings_history"})
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) createUser(ctx context.Context, user models.UserRequest) models.User {
	s.T().Helper()
	newPhone := uuid.New().String()
	user.Phone = &newPhone
	result := models.User{}
	resp := s.sendRequest(ctx, http.MethodPost, "/api/v1/users", user, &result)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	return result
}

func (s *IntegrationTestSuite) getToken(ctx context.Context, phone, password string) string {
	s.T().Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+`/api/v1/login`, nil)
	s.Require().NoError(err)
	req.SetBasicAuth(phone, password)
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	var token models.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	s.Require().NoError(err)
	return token.Token
}

func (s *IntegrationTestSuite) createMeeting(ctx context.Context, meeting models.MeetingRequest) (models.Meeting, string) {
	s.T().Helper()
	testUser1 := s.createUser(ctx, user)
	testUser2 := s.createUser(ctx, user)
	meeting.Manager = &testUser1.ID
	meeting.Client = &testUser2.ID
	token := s.getToken(ctx, testUser1.Phone, *user.Password)
	reqBody, err := json.Marshal(meeting)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+"/api/v1/meetings", bytes.NewReader(reqBody))
	s.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	var respMeeting models.Meeting
	err = json.NewDecoder(resp.Body).Decode(&respMeeting)
	s.Require().NoError(err)
	return respMeeting, token
}

func (s *IntegrationTestSuite) TestCreateUser() {
	ctx := context.Background()
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
	s.Require().Equal(*user.LastName, respUser.LastName)
	s.Require().Equal(*user.FirstName, respUser.FirstName)
	s.Require().Equal(*user.Email, respUser.Email)
	s.Require().Equal(*user.Phone, respUser.Phone)
	s.Require().Equal(respUser.CreatedAt, respUser.UpdatedAt)
	var cnt int
	err = s.store.QueryRow(ctx, `SELECT count(*) FROM users_history WHERE user_id = $1`, respUser.ID).Scan(&cnt)
	s.Require().NoError(err)
	s.Require().Equal(1, cnt)
}

func (s *IntegrationTestSuite) TestGetUser() {
	ctx := context.Background()
	testUser := s.createUser(ctx, user)
	token := s.getToken(ctx, testUser.Phone, *user.Password)

	s.Run("get user", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
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
		s.Require().Equal(testUser.ID, respUser.ID)
		s.Require().Equal(testUser.LastName, respUser.LastName)
		s.Require().Equal(testUser.FirstName, respUser.FirstName)
		s.Require().Equal(testUser.Email, respUser.Email)
		s.Require().Equal(testUser.Phone, respUser.Phone)
	})

	s.Run("get user not found", func() {
		testUser.ID = 0
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err getting user (id %d) from store: %v", testUser.ID, pgstore.ErrUserNotFound), respError.Error)
	})
}

func (s *IntegrationTestSuite) TestUpdateUser() {
	ctx := context.Background()

	var (
		lastName = "Updated"
		password = "Boop!"
		phone    = "+1778979900"
		role     = "coach"
	)

	data := models.UserRequest{
		LastName:  &lastName,
		FirstName: &password,
		Phone:     &phone,
		Role:      &role,
	}

	testUser := s.createUser(ctx, user)
	token := s.getToken(ctx, testUser.Phone, *user.Password)
	reqBody, err := json.Marshal(data)
	s.Require().NoError(err)

	s.Run("update user", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
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
		s.Require().Equal(testUser.ID, respUser.ID)
		s.Require().Equal(*data.LastName, respUser.LastName)
		s.Require().Equal(*data.FirstName, respUser.FirstName)
		s.Require().Equal(*data.Phone, respUser.Phone)
	})

	s.Run("update user not found", func() {
		testUser.ID = 0
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err updating user (id %d) from store: %v", testUser.ID, pgstore.ErrUserNotFound), respError.Error)
	})
}

func (s *IntegrationTestSuite) TestDeleteUser() {
	ctx := context.Background()
	testUser := s.createUser(ctx, user)
	token := s.getToken(ctx, testUser.Phone, *user.Password)

	s.Run("delete user", func() {
		reqBody, err := json.Marshal(user)
		s.Require().NoError(err)
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
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
		s.Require().Equal(testUser.ID, respUser.ID)
		s.Require().Equal(testUser.LastName, respUser.LastName)
		s.Require().Equal(testUser.FirstName, respUser.FirstName)
		s.Require().Equal(true, respUser.Deleted)
		s.Require().Equal(testUser.Email, respUser.Email)
	})

	s.Run("delete user not found", func() {
		testUser.ID = 0
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/users/"+strconv.Itoa(testUser.ID), nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err deleting user (id %d) from store: %v", testUser.ID, pgstore.ErrUserNotFound), respError.Error)
	})
}

func (s *IntegrationTestSuite) TestCreateMeeting() {
	ctx := context.Background()
	testUser1 := s.createUser(ctx, user)
	testUser2 := s.createUser(ctx, user)
	meeting.Manager = &testUser1.ID
	meeting.Client = &testUser2.ID
	user.ID = &testUser1.ID
	token := s.getToken(ctx, testUser1.Phone, *user.Password)
	reqBody, err := json.Marshal(meeting)
	s.Require().NoError(err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+"/api/v1/meetings", bytes.NewReader(reqBody))
	s.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	var respMeeting models.Meeting
	err = json.NewDecoder(resp.Body).Decode(&respMeeting)
	s.Require().NoError(err)
	s.Require().Equal(*meeting.Manager, respMeeting.Manager)
	s.Require().Equal(*meeting.Client, respMeeting.Client)
	s.Require().Equal(*meeting.StartTime, respMeeting.StartTime.UTC())
	s.Require().Equal(*meeting.EndTime, respMeeting.EndTime.UTC())
}

func (s *IntegrationTestSuite) TestGetMeeting() {
	ctx := context.Background()
	newMeeting, token := s.createMeeting(ctx, meeting)
	s.Run("get meeting", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/meetings/"+strconv.Itoa(newMeeting.ID), nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respMeeting models.Meeting
		err = json.NewDecoder(resp.Body).Decode(&respMeeting)
		s.Require().NoError(err)

	})

	s.Run("not found meeting", func() {
		id := 0
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL+"/api/v1/meetings/0", nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err getting meeting (id %d) from store: %v", id, pgstore.ErrMeetingNotFound), respError.Error)
	})
}

func (s *IntegrationTestSuite) TestUpdateMeeting() {
	ctx := context.Background()
	data := models.Meeting{
		Manager:   2,
		StartTime: time.Date(2023, 1, 10, 12, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2023, 1, 10, 24, 0, 0, 0, time.UTC),
		Client:    1,
	}
	newMeeting, token := s.createMeeting(ctx, meeting)
	reqBody, err := json.Marshal(data)
	s.Require().NoError(err)

	s.Run("update meeting", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/meetings/"+strconv.Itoa(newMeeting.ID), bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respMeeting models.Meeting
		err = json.NewDecoder(resp.Body).Decode(&respMeeting)
		s.Require().NoError(err)
		s.Require().Equal(newMeeting.ID, respMeeting.ID)
		s.Require().Equal(data.Manager, respMeeting.Manager)
		s.Require().Equal(data.Client, respMeeting.Client)
		s.Require().Equal(data.StartTime, respMeeting.StartTime.UTC())
		s.Require().Equal(data.EndTime, respMeeting.EndTime.UTC())
	})

	s.Run("not found meeting", func() {
		id := 0
		req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL+"/api/v1/meetings/0", bytes.NewReader(reqBody))
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err updating meeting (id %d) from store: %v", id, pgstore.ErrMeetingNotFound), respError.Error)
	})
}

func (s *IntegrationTestSuite) TestDeleteMeeting() {
	ctx := context.Background()
	newMeeting, token := s.createMeeting(ctx, meeting)

	s.Run("delete meeting", func() {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/meetings/"+strconv.Itoa(newMeeting.ID), nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		var respMeeting models.Meeting
		err = json.NewDecoder(resp.Body).Decode(&respMeeting)
		s.Require().NoError(err)
		s.Require().Equal(newMeeting.ID, respMeeting.ID)
		s.Require().Equal(newMeeting.Manager, respMeeting.Manager)
		s.Require().Equal(newMeeting.Client, respMeeting.Client)
		s.Require().Equal(newMeeting.StartTime.UTC(), respMeeting.StartTime.UTC())
		s.Require().Equal(newMeeting.EndTime.UTC(), respMeeting.EndTime.UTC())
	})

	s.Run("not found meeting", func() {
		id := 0
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL+"/api/v1/meetings/0", nil)
		s.Require().NoError(err)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)
		defer func() {
			err = resp.Body.Close()
			s.Require().NoError(err)
		}()
		s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		var respError errResp
		err = json.NewDecoder(resp.Body).Decode(&respError)
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("err deleting meeting (id %d) from store: %v", id, pgstore.ErrMeetingNotFound), respError.Error)
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

func (s *IntegrationTestSuite) TestGenerateHashFromPassword() {
	hash, err := bcrypt.GenerateFromPassword([]byte("jopa"), 0)
	s.Require().NoError(err)
	s.T().Log(string(hash))
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
