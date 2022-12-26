package store

import (
	"context"

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

func (s *Store) GetUsersSQL(ctx context.Context) ([]models.User, error) {
	var users []models.User
	var tmp models.User
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM users`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			s.log.Warnf("err during closing rows: %v", err)
		}
	}()
	for rows.Next() {
		if err = rows.Scan(&tmp.ID, &tmp.LastName, &tmp.FirstName, &tmp.MiddleName); err != nil {
			return nil, err
		}
		users = append(users, tmp)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
