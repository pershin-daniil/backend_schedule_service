package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"strings"
)

//go:embed migrations
var migrations embed.FS

const retries = 3

type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

var ErrUserNotFound = fmt.Errorf("user not found")
var ErrMeetingNotFound = fmt.Errorf("meeting not found")

func NewStore(ctx context.Context, log *logrus.Logger, dsn string) (*Store, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{
		log: log.WithField("component", "pgstore"),
		db:  db,
	}, nil
}

func (s *Store) Migrate(direction migrate.MigrationDirection) error {
	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, er := migrations.ReadDir(path)
			if er != nil {
				return nil, er
			}
			entries := make([]string, 0)
			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()
	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}
	_, err := migrate.Exec(s.db.DB, "postgres", asset, direction)
	return err
}

func (s *Store) GetUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.SelectContext(ctx, &users, `SELECT * FROM users`); err != nil {
			continue
		}
		return users, nil
	}
	return nil, err
}

func (s *Store) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	var createdUser models.User
	query := `
INSERT INTO users (last_name, first_name)
VALUES ($1, $2)
RETURNING id, last_name, first_name;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.QueryRowxContext(ctx, query, user.LastName, user.FirstName).
			Scan(&createdUser.ID, &createdUser.LastName, &createdUser.FirstName); err != nil {
			continue
		}
		return createdUser, nil
	}
	return models.User{}, fmt.Errorf("err getting users: %w", err)
}

func (s *Store) GetUser(ctx context.Context, id int) (models.User, error) {
	var user models.User
	query := `
SELECT * FROM users
WHERE id = $1;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &user, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, ErrUserNotFound
		case err != nil:
			continue
		}
		return user, nil
	}
	return models.User{}, fmt.Errorf("err getting user %d: %w", id, err)
}

func (s *Store) UpdateUser(ctx context.Context, id int, user models.User) (models.User, error) {
	var updatedUser models.User
	query := `
UPDATE users
    SET last_name = $2,
    first_name = $3
WHERE id = $1
RETURNING *;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &updatedUser, query, id, user.LastName, user.FirstName)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, ErrUserNotFound
		case err != nil:
			continue
		}
		return updatedUser, nil
	}
	return models.User{}, fmt.Errorf("err updating user %d: %w", id, err)
}

func (s *Store) DeleteUser(ctx context.Context, id int) (models.User, error) {
	var deletedUser models.User
	query := `
DELETE FROM users
WHERE id = $1
RETURNING *;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &deletedUser, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, ErrUserNotFound
		case err != nil:
			continue
		}
		return deletedUser, err
	}
	return models.User{}, fmt.Errorf("err deleting user %d: %w", id, err)
}

func (s *Store) CreateMeeting(ctx context.Context, meeting models.Meeting) (models.Meeting, error) {
	var newMeeting models.Meeting
	query := `
INSERT INTO meetings (manager, start_at, end_at, client)
VALUES ($1, $2, $3, $4)
RETURNING id, manager, start_at, end_at, client;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.QueryRowxContext(ctx, query, meeting.Manager, meeting.StartTime, meeting.EndTime, meeting.Client).
			Scan(&newMeeting.ID, &newMeeting.Manager, &newMeeting.StartTime, &newMeeting.EndTime, &newMeeting.Client); err != nil {
			continue
		}
		return newMeeting, err
	}
	return models.Meeting{}, err
}

func (s *Store) GetMeetings(ctx context.Context) ([]models.Meeting, error) {
	var meetings []models.Meeting
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.SelectContext(ctx, &meetings, `SELECT * FROM meetings`); err != nil {
			continue
		}
		return meetings, nil
	}
	return nil, err
}

func (s *Store) GetMeeting(ctx context.Context, id int) (models.Meeting, error) {
	var meeting models.Meeting
	query := `
SELECT * FROM meetings
WHERE id = $1;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &meeting, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.Meeting{}, ErrMeetingNotFound
		case err != nil:
			continue
		}
		return meeting, nil
	}
	return models.Meeting{}, fmt.Errorf("err getting meeting %d: %w", id, err)
}

func (s *Store) UpdateMeeting(ctx context.Context, id int, meeting models.Meeting) (models.Meeting, error) {
	var updatedMeeting models.Meeting
	query := `
UPDATE meetings
SET manager = $2,
	start_at = $3,
	end_at = $4,
	client = $5
WHERE id = $1
RETURNING *`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &updatedMeeting, query, id, meeting.Manager, meeting.StartTime, meeting.EndTime, meeting.Client)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.Meeting{}, ErrMeetingNotFound
		case err != nil:
			continue
		}
		return updatedMeeting, nil
	}
	return models.Meeting{}, fmt.Errorf("err updating meeting %d: %w", id, err)
}

func (s *Store) DeleteMeeting(ctx context.Context, id int) (models.Meeting, error) {
	var deletedMeeting models.Meeting
	query := `
DELETE FROM meetings
WHERE id = $1
RETURNING *;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &deletedMeeting, query, id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.Meeting{}, ErrMeetingNotFound
		case err != nil:
			continue
		}
		return deletedMeeting, nil
	}
	return models.Meeting{}, fmt.Errorf("err deleting meeting %d: %w", id, err)
}

func (s *Store) ResetTables(ctx context.Context, tables []string) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE `+strings.Join(tables, `, `))
	for _, table := range tables {
		_, err = s.db.ExecContext(ctx, fmt.Sprintf(`ALTER SEQUENCE %s_id_seq RESTART`, table))
		if err != nil {
			return err
		}
	}
	return err
}
