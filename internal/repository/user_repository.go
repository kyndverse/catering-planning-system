package repository

import (
	"database/sql"
	"golang-web/internal/domain/entity"

	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	Log *logrus.Logger
}

func NewUserRepository(logger *logrus.Logger) *UserRepository {
	return &UserRepository{
		Log: logger,
	}
}

func (r *UserRepository) FindByEmail(tx *sql.Tx, user *entity.User) error {
	query := `SELECT id, name, email, password, role, created_at FROM users WHERE email=$1 LIMIT 1`

	return tx.QueryRow(query, user.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
}
