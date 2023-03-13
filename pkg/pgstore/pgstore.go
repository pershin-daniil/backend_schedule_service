package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pershin-daniil/TimeSlots/pkg/metrics"

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
	ErrUserExists      = fmt.Errorf("user already exists")
)

func New(ctx context.Context, log *logrus.Logger, dsn string) (*Store, error) {
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
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetUsers").Observe(time.Since(started).Seconds())
	}()
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
	metrics.PgErrCount.WithLabelValues("GetUsers").Inc()

	return nil, fmt.Errorf("get users failed: %w", err)
}

func (s *Store) CreateUser(ctx context.Context, user models.UserRequest) (models.User, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetUser").Observe(time.Since(started).Seconds())
	}()

	tx, err := s.db.Beginx()
	if err != nil {
		return models.User{}, fmt.Errorf("create user failed: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("create user failed: %v", err)
		}
	}()

	exists, err := s.userExists(ctx, tx, user)
	if err != nil {
		return models.User{}, err
	}
	if exists {
		return models.User{}, ErrUserExists
	}
	var createdUser models.User
	query := `
INSERT INTO users (last_name, first_name, phone, email, password_hash, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, last_name, first_name, phone, COALESCE(email, '') AS email, updated_at, created_at, password_hash, role;`

	for i := 0; i < retries; i++ {
		if err = s.db.GetContext(ctx, &createdUser, query, user.LastName, user.FirstName, user.Phone, user.Email, user.PasswordHash, user.Role); err != nil {
			continue
		}
		return createdUser, nil
	}
	metrics.PgErrCount.WithLabelValues("GetUser").Inc()

	return models.User{}, fmt.Errorf("create user failed: %w", err)
}

type requester interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (s *Store) userExists(ctx context.Context, requester requester, user models.UserRequest) (bool, error) {
	query := `
SELECT TRUE FROM users
WHERE phone=$1 AND NOT deleted;`
	var exists bool
	var err error

	for i := 0; i < retries; i++ {
		err = requester.GetContext(ctx, &exists, query, user.Phone)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		case err != nil:
			continue
		}
		return true, nil
	}
	metrics.PgErrCount.WithLabelValues("userExists").Inc()

	return false, err
}

func (s *Store) GetUserByPhone(ctx context.Context, phone string) (models.User, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetUserByPhone").Observe(time.Since(started).Seconds())
	}()

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
	metrics.PgErrCount.WithLabelValues("GetUserByPhone").Inc()

	return models.User{}, fmt.Errorf("get user by phone (%s) faild: %w", phone, err)
}

func (s *Store) GetUser(ctx context.Context, id int) (models.User, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetUser").Observe(time.Since(started).Seconds())
	}()

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
	metrics.PgErrCount.WithLabelValues("GetUser").Inc()

	return models.User{}, fmt.Errorf("get user %d faild: %w", id, err)
}

func (s *Store) UpdateUser(ctx context.Context, id int, user models.UserRequest) (models.User, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("UpdateUser").Observe(time.Since(started).Seconds())
	}()

	tx, err := s.db.Beginx()
	if err != nil {
		return models.User{}, fmt.Errorf("open transaction faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("rollback transaction faild: %v", err)
		}
	}()

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
	for i := 0; i < retries; i++ {
		err = tx.GetContext(ctx, &updatedUser, query.String(), args...)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, ErrUserNotFound
		case err != nil:
			continue
		}
		return updatedUser, nil
	}
	metrics.PgErrCount.WithLabelValues("UpdateUser").Inc()

	return models.User{}, fmt.Errorf("update user %d faild: %w", id, err)
}

func (s *Store) DeleteUser(ctx context.Context, id int) (models.User, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("DeleteUser").Observe(time.Since(started).Seconds())
	}()

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
	metrics.PgErrCount.WithLabelValues("DeleteUser").Inc()

	return models.User{}, fmt.Errorf("delete user %d faild: %w", id, err)
}

func (s *Store) CreateMeeting(ctx context.Context, meeting models.MeetingRequest) (models.Meeting, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("CreateMeeting").Observe(time.Since(started).Seconds())
	}()

	var newMeeting models.Meeting
	query := `
INSERT INTO meetings (manager, start_at, end_at, client)
VALUES ($1, $2, $3, $4)
RETURNING id, manager, start_at, end_at, client, updated_at, created_at;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.GetContext(ctx, &newMeeting, query, meeting.Manager, meeting.StartTime, meeting.EndTime, meeting.Client); err != nil {
			continue
		}
		return newMeeting, err
	}
	metrics.PgErrCount.WithLabelValues("CreateMeeting").Inc()

	return models.Meeting{}, fmt.Errorf("create meeting faild: %w", err)
}

func (s *Store) GetMeetings(ctx context.Context) ([]models.Meeting, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetMeetings").Observe(time.Since(started).Seconds())
	}()

	var meetings []models.Meeting
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.SelectContext(ctx, &meetings, `SELECT id, manager, start_at, end_at, client, updated_at, created_at FROM meetings`); err != nil {
			continue
		}
		return meetings, nil
	}
	metrics.PgErrCount.WithLabelValues("GetMeetings").Inc()

	return nil, fmt.Errorf("get meetings faild: %w", err)
}

func (s *Store) GetMeeting(ctx context.Context, id int) (models.Meeting, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetMeeting").Observe(time.Since(started).Seconds())
	}()

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
	metrics.PgErrCount.WithLabelValues("GetMeeting").Inc()

	return models.Meeting{}, fmt.Errorf("get meeting %d faild: %w", id, err)
}

func (s *Store) UpdateMeeting(ctx context.Context, id int, meeting models.MeetingRequest) (models.Meeting, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("UpdateMeeting").Observe(time.Since(started).Seconds())
	}()

	tx, err := s.db.Beginx()
	if err != nil {
		return models.Meeting{}, fmt.Errorf("open transaction faild: %w", err)
	}
	defer func() {
		if err = tx.Rollback(); err != nil && errors.Is(err, sql.ErrTxDone) {
			s.log.Warnf("rollback transaction faild: %v", err)
		}
	}()

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
	for i := 0; i < retries; i++ {
		err = tx.GetContext(ctx, &updatedMeeting, query.String(), args...)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.Meeting{}, ErrMeetingNotFound
		case err != nil:
			continue
		}
		return updatedMeeting, nil
	}
	metrics.PgErrCount.WithLabelValues("UpdateMeeting").Inc()

	return models.Meeting{}, fmt.Errorf("update meeting %d faild: %w", id, err)
}

func (s *Store) DeleteMeeting(ctx context.Context, id int) (models.Meeting, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("DeleteMeeting").Observe(time.Since(started).Seconds())
	}()

	var deletedMeeting models.Meeting
	query := `
DELETE FROM meetings
WHERE id = $1
RETURNING id, manager, start_at, end_at, client, updated_at, created_at;`
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
	metrics.PgErrCount.WithLabelValues("DeleteMeeting").Inc()

	return models.Meeting{}, fmt.Errorf("delete meeting %d faild: %w", id, err)
}

func (s *Store) UsersWithMeetings(ctx context.Context) ([]models.UserNotify, error) {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("GetUsersForNotification").Observe(time.Since(started).Seconds())
	}()

	var result []models.UserNotify
	query := `
SELECT users.id AS user_id, m.id AS meeting_id, notified, last_name, first_name, start_at FROM users
JOIN meetings m on users.id = m.client
WHERE now() < start_at + users.notification
AND NOT notified`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.SelectContext(ctx, &result, query); err != nil {
			continue
		}
		return result, nil
	}
	metrics.PgErrCount.WithLabelValues("UsersWithMeeting").Inc()

	return nil, fmt.Errorf("get users with meetings faild: %w", err)
}

func (s *Store) SwitchNotificationStatus(ctx context.Context, meetingID int) error {
	started := time.Now()
	defer func() {
		metrics.PgDuration.WithLabelValues("SwitchNotificationStatus").Observe(time.Since(started).Seconds())
	}()

	var result bool
	query := `
UPDATE meetings SET
notified = true
WHERE id = $1
RETURNING TRUE;`
	var err error
	for i := 0; i < retries; i++ {
		if err = s.db.GetContext(ctx, &result, query, meetingID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if result || errors.Is(err, sql.ErrNoRows) {
			return nil
		}
	}
	metrics.PgErrCount.WithLabelValues("SwitchNotificationStatus").Inc()

	return fmt.Errorf("switch notification status faild: %w", err)
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
