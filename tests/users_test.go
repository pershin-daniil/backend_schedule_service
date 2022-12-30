package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/stretchr/testify/require"
)

var testURL = "http://localhost:8080"

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	user := models.User{
		LastName:    "TestLN",
		FirstName:   "TestFN",
		PhoneNumber: 1234567890,
	}
	reqBody, err := json.Marshal(user)
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL+"/api/v1/createUser", bytes.NewReader(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var respUser models.User
	err = json.NewDecoder(resp.Body).Decode(&respUser)
	require.NoError(t, err)
	require.Equal(t, user.LastName, respUser.LastName)
	require.Equal(t, user.FirstName, respUser.FirstName)
	require.Equal(t, user.PhoneNumber, respUser.PhoneNumber)
}
