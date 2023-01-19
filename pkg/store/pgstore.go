package store

import (
	"context"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pershin-daniil/TimeSlots/pkg/models"
	"github.com/sirupsen/logrus"
)

type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

func NewStore(log *logrus.Logger, dsn string) (*Store, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{
		log: log.WithField("component", "store"),
		db:  db,
	}, nil
}
func (s *Store) GetUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := s.db.SelectContext(ctx, &users, `SELECT * FROM users`)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	var result models.User
	query := `
INSERT INTO users (last_name, first_name)
 VALUES ($1, $2)
  RETURNING user_id, last_name, first_name`
	err := s.db.QueryRowxContext(ctx, query, user.LastName, user.FirstName).
		Scan(&result.ID, &result.LastName, &result.FirstName)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *Store) ReadUser(ctx context.Context, id string) ([]models.User, error) {
	var user []models.User
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	query := `
SELECT * FROM users
WHERE user_id = $1
`
	err = s.db.SelectContext(ctx, &user, query, idInt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, id string, user models.User) ([]models.User, error) {
	var result []models.User
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	query := `
UPDATE users
    SET last_name = $2,
    first_name = $3
WHERE user_id = $1
RETURNING *`
	err = s.db.SelectContext(ctx, &result, query, idInt, user.LastName, user.FirstName)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) DeleteUser(ctx context.Context, id string) ([]models.User, error) {
	var deletedUser []models.User
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	query := `
DELETE FROM users
WHERE user_id = $1
RETURNING *;
`
	err = s.db.SelectContext(ctx, &deletedUser, query, idInt)
	return deletedUser, err
}

func (s *Store) TruncateTable(ctx context.Context, table string) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE TABLE `+table)
	return err
}
