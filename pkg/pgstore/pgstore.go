package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations
var migrations embed.FS

const retries = 3

type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

var (
	ErrUserNotFound    = fmt.Errorf("user not found")
	ErrMeetingNotFound = fmt.Errorf("meeting not found")
)

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
	if err != nil {
		return fmt.Errorf("err migrating: %w", err)
	}
	s.log.Infof("migration success")
	return nil
}

func (s *Store) GetUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	var err error
	query := `SELECT id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at FROM users
WHERE NOT deleted;`
	for i := 0; i < retries; i++ {
		if err = s.db.SelectContext(ctx, &users, query); err != nil {
			continue
		}
		return users, nil
	}
	return nil, err
}

func (s *Store) CreateUser(ctx context.Context, user models.UserRequest) (models.User, error) {
	var createdUser models.User
	if ok := s.isUserExists(ctx, user); !ok {
		return models.User{}, fmt.Errorf("err user already exists")
	}
	query := `
INSERT INTO users (last_name, first_name, phone, email, password_hash, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at, password_hash, role;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.QueryRowxContext(ctx, query, user.LastName, user.FirstName, user.Phone, user.Email, user.PasswordHash, user.Role).
			StructScan(&createdUser); err != nil {
			continue
		}
		return createdUser, nil
	}
	return models.User{}, fmt.Errorf("err creating users: %w", err)
}

func (s *Store) isUserExists(ctx context.Context, user models.UserRequest) bool {
	query := `
SELECT last_name, first_name, phone FROM users
WHERE last_name=$1 AND first_name=$2 OR phone=$3 AND NOT deleted;`
	if err := s.db.GetContext(ctx, models.User{}, query, user.LastName, user.FirstName, user.Phone); errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return true
}

func (s *Store) GetUserByPhone(ctx context.Context, phone string) (models.User, error) {
	var user models.User
	query := `
SELECT id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at, phone, password_hash, role
FROM users
WHERE phone = $1 AND NOT deleted;`
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &user, query, phone)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, ErrUserNotFound
		case err != nil:
			continue
		}
		return user, nil
	}
	return models.User{}, fmt.Errorf("err getting user by phone (%s): %w", phone, err)
}

func (s *Store) GetUser(ctx context.Context, id int) (models.User, error) {
	var user models.User
	query := `
SELECT id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at FROM users
WHERE id = $1 AND NOT deleted;`
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

func (s *Store) UpdateUser(ctx context.Context, id int, user models.UserRequest) (models.User, error) {
	var updatedUser models.User
	var args []interface{}
	var query strings.Builder
	query.WriteString(`UPDATE users SET` + ` `)
	if user.LastName != nil {
		args = append(args, *user.LastName)
		query.WriteString(`last_name = $` + fmt.Sprint(len(args)) + `, `)
	}
	if user.FirstName != nil {
		args = append(args, *user.FirstName)
		query.WriteString(`first_name = $` + fmt.Sprint(len(args)) + `, `)
	}
	if user.Phone != nil {
		args = append(args, *user.Phone)
		query.WriteString(`phone = $` + fmt.Sprint(len(args)) + `, `)
	}
	if user.Email != nil {
		args = append(args, *user.Email)
		query.WriteString(`email = $` + fmt.Sprint(len(args)) + `, `)
	}
	args = append(args, id)
	query.WriteString(fmt.Sprintf(` updated_at = NOW() WHERE id = $%d
RETURNING id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at;`, len(args)))
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &updatedUser, query.String(), args...)
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
UPDATE users
SET deleted = true
WHERE id = $1
RETURNING id, last_name, first_name, phone, COALESCE(email, '') AS email, deleted, updated_at, created_at;`
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

func (s *Store) CreateMeeting(ctx context.Context, meeting models.MeetingRequest) (models.Meeting, error) {
	var newMeeting models.Meeting
	query := `
INSERT INTO meetings (manager, start_at, end_at, client)
VALUES ($1, $2, $3, $4)
RETURNING id, manager, start_at, end_at, client, updated_at, created_at;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.QueryRowxContext(ctx, query, meeting.Manager, meeting.StartTime, meeting.EndTime, meeting.Client).
			StructScan(&newMeeting); err != nil {
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
		if err = s.db.SelectContext(ctx, &meetings, `SELECT id, manager, start_at, end_at, client, updated_at, created_at FROM meetings`); err != nil {
			continue
		}
		return meetings, nil
	}
	return nil, err
}

func (s *Store) GetMeeting(ctx context.Context, id int) (models.Meeting, error) {
	var meeting models.Meeting
	query := `
SELECT id, manager, start_at, end_at, client, updated_at, created_at FROM meetings
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

func (s *Store) UpdateMeeting(ctx context.Context, id int, meeting models.MeetingRequest) (models.Meeting, error) {
	var updatedMeeting models.Meeting
	var args []interface{}
	var query strings.Builder
	query.WriteString(`UPDATE meetings SET` + ` `)
	if meeting.Manager != nil {
		args = append(args, *meeting.Manager)
		query.WriteString(`manager = $` + fmt.Sprint(len(args)) + `, `)
	}
	if meeting.StartTime != nil {
		args = append(args, *meeting.StartTime)
		query.WriteString(`start_at = $` + fmt.Sprint(len(args)) + `, `)
	}
	if meeting.EndTime != nil {
		args = append(args, *meeting.EndTime)
		query.WriteString(`end_at = $` + fmt.Sprint(len(args)) + `, `)
	}
	if meeting.Client != nil {
		args = append(args, *meeting.Client)
		query.WriteString(`client = $` + fmt.Sprint(len(args)) + `, `)
	}
	args = append(args, id)
	query.WriteString(fmt.Sprintf(` updated_at = NOW() WHERE id = $%d
RETURNING id, manager, start_at, end_at, client, updated_at, created_at;`, len(args)))
	var err error
	for i := 0; i < retries; i++ {
		err = s.db.GetContext(ctx, &updatedMeeting, query.String(), args...)
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
	// TODO: don't delete History
	var deletedHistory interface{}
	query := `
DELETE FROM meetings
WHERE id = $1
RETURNING id, manager, start_at, end_at, client, updated_at, created_at;`
	var err error
	for i := 0; i < retries; i++ {
		_ = s.db.GetContext(ctx, &deletedHistory, `
DELETE FROM meetings_history
WHERE meetings_id = $1
RETURNING NULL;
`, id)
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
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE`+` `+strings.Join(tables, `, `))
	for _, table := range tables {
		_, err = s.db.ExecContext(ctx, fmt.Sprintf(`ALTER SEQUENCE %s_id_seq RESTART`, table))
		if err != nil {
			return err
		}
	}
	return err
}

func (s *Store) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *Store) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}
