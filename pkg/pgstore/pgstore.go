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
)

//go:embed migrations
var migrations embed.FS

const retries = 3

type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

var ErrUserNotFound = errors.New("user not found")

func NewStore(ctx context.Context, log *logrus.Logger, dsn string) (*Store, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{
		log: log.WithField("component", "store"),
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
RETURNING user_id, last_name, first_name;`
	err := s.db.QueryRowxContext(ctx, query, user.LastName, user.FirstName).
		Scan(&createdUser.ID, &createdUser.LastName, &createdUser.FirstName)
	if err != nil {
		return models.User{}, err
	}
	return createdUser, nil
}

func (s *Store) GetUser(ctx context.Context, id int) (models.User, error) {
	var user models.User
	query := `
SELECT * FROM users
WHERE user_id = $1;`
	err := s.db.GetContext(ctx, &user, query, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.User{}, ErrUserNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("err getting user %d: %w", id, err)
	}
	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, id int, user models.User) (models.User, error) {
	var updatedUser models.User
	query := `
UPDATE users
    SET last_name = $2,
    first_name = $3
WHERE user_id = $1
RETURNING *;`
	err := s.db.GetContext(ctx, &updatedUser, query, id, user.LastName, user.FirstName)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.User{}, ErrUserNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("err updating user %d: %w", id, err)
	}
	return updatedUser, nil
}

func (s *Store) DeleteUser(ctx context.Context, id int) (models.User, error) {
	var deletedUser models.User
	query := `
DELETE FROM users
WHERE user_id = $1
RETURNING *;`
	err := s.db.GetContext(ctx, &deletedUser, query, id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.User{}, ErrUserNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("err deleting user %d: %w", id, err)
	}
	return deletedUser, err
}

func (s *Store) TruncateTable(ctx context.Context, table string) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE `+table)
	return err
}
