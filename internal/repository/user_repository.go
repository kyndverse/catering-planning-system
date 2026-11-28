package repository

import (
	"database/sql"
	"golang-web/internal/entity"

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

func (r *UserRepository) FindAll(tx *sql.Tx) ([]entity.User, error) {
	query := `SELECT id, name, email, role, created_at FROM users ORDER BY id`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		u := entity.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *UserRepository) FindOneByID(tx *sql.Tx, user *entity.User) error {
	query := `SELECT id, name, email, role, created_at FROM users WHERE id=$1 LIMIT 1`

	return tx.QueryRow(query, user.ID).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
}

func (r *UserRepository) ExistsByEmail(tx *sql.Tx, email string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`

	var exists bool
	if err := tx.QueryRow(query, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) Create(tx *sql.Tx, user *entity.User) error {
	query := `
        INSERT INTO users (name, email, password, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `

	return tx.QueryRow(query, user.Name, user.Email, user.Password, user.Role).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) Update(tx *sql.Tx, user *entity.User) error {
	query := `UPDATE users SET name=$1, role=$2 WHERE id=$3 RETURNING id, name, email, role, created_at`

	return tx.QueryRow(query, user.Name, user.Role, user.ID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt,
	)
}

func (r *UserRepository) UpdatePassword(tx *sql.Tx, user *entity.User) error {
	query := `UPDATE users SET password=$1 WHERE id=$2`

	_, err := tx.Exec(query, user.Password, user.ID)
	return err
}

func (r *UserRepository) Delete(tx *sql.Tx, user *entity.User) error {
	query := `DELETE FROM users WHERE id=$1`

	_, err := tx.Exec(query, user.ID)
	return err
}
