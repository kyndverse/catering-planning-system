package service

import (
	"context"
	"database/sql"
	"errors"
	"golang-web/internal/auth"
	"golang-web/internal/domain/entity"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/validator"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Log               *logrus.Logger
	DB                *sql.DB
	Validator         *validator.Validator
	UserRepository    *repository.UserRepository
	SessionRepository *repository.SessionRepository
}

func NewAuthService(logger *logrus.Logger, db *sql.DB, validator *validator.Validator, userRepository *repository.UserRepository, sessionRepository *repository.SessionRepository) *AuthService {
	return &AuthService{
		Log:               logger,
		DB:                db,
		Validator:         validator,
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
	}
}

func (s *AuthService) Login(ctx context.Context, request *model.LoginRequest, oldSessionID string) (string, validator.ValidatorErrors, error) {
	if fieldErrors := s.Validator.ValidateStruct(request); fieldErrors != nil {
		return "", fieldErrors, nil
	}

	var formError validator.ValidatorErrors

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return "", nil, err
	}
	defer tx.Rollback()

	user := new(entity.User)
	user.Email = request.Email
	if err := s.UserRepository.FindByEmail(tx, user); err != nil {
		if user.Password == "" {
			formError = validator.ValidatorErrors{
				"Error": "Email atau password salah!",
			}

			return "", formError, nil
		}

		s.Log.WithError(err).Error("failed to find user by email")
		return "", nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			s.Log.WithError(err).Error("failed to compare and hash password")
			return "", nil, err
		}
		s.Log.Warn("login failed: invalid_password")

		formError = validator.ValidatorErrors{
			"Error": "Email atau password salah!",
		}

		return "", formError, nil
	}

	if oldSessionID != "" {
		if err := s.SessionRepository.Delete(ctx, oldSessionID); err != nil {
			s.Log.WithError(err).Error("failed to delete session")
			return "", nil, err
		}
	}

	sessionID, err := auth.GenerateToken()
	if err != nil {
		s.Log.WithError(err).Error("failed to generate token")
		return "", nil, err
	}

	const sessionDuration = 24 * time.Hour
	now := time.Now()

	session := &entity.Session{
		ID:        sessionID,
		Role:      user.Role,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionDuration),
	}

	if err := s.SessionRepository.Create(ctx, session, sessionDuration); err != nil {
		s.Log.WithError(err).Error("failed to create session")
		return "", nil, err
	}

	return sessionID, nil, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {

	if sessionID == "" {
		return nil
	}

	if err := s.SessionRepository.Delete(ctx, sessionID); err != nil {
		s.Log.WithError(err).Error("failed to delete session")
		return err
	}

	return nil
}
