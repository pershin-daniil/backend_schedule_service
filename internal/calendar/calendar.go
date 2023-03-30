package calendar

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

//go:embed credentials
var credentials embed.FS

type Calendar struct {
	log *logrus.Entry
	srv *calendar.Service
}

func New(ctx context.Context, log *logrus.Logger) *Calendar {
	srv := calendarService(ctx)
	return &Calendar{
		log: log.WithField("module", "calendar"),
		srv: srv,
	}
}

func (c *Calendar) Events() []models.Event {
	t := time.Now().Format(time.RFC3339)
	events, err := c.srv.Events.List("1504342299bb251fb7f535b0707b29fbc9b9e4f0d6fba28f0cb8f7f8e1cd3355@group.calendar.google.com").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		c.log.Panicf("Unable to retrieve next ten of the user's events: %v", err)
	}

	if len(events.Items) == 0 {
		return nil
	}

	result := make([]models.Event, 0, len(events.Items))

	for _, item := range events.Items {
		event := models.Event{
			ID:          item.Id,
			Title:       item.Summary,
			Description: item.Description,
			Start:       item.Start.DateTime,
			End:         item.End.DateTime,
			Created:     item.Created,
			Updated:     item.Updated,
			Status:      item.Status,
		}
		result = append(result, event)
	}

	return result
}

func calendarService(ctx context.Context) *calendar.Service {
	b, err := credentials.ReadFile("credentials/credentials.json")
	if err != nil {
		log.Panicf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Panicf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Panicf("Unable to retrieve Calendar client: %v", err)
	}
	return srv
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Panicf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Panicf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Panicf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
