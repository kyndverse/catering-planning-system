package service

import (
	"context"
	"database/sql"

	"golang-web/internal/entity"
	"golang-web/internal/model"
	"golang-web/internal/repository"
	"golang-web/internal/validator"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Log            *logrus.Logger
	DB             *sql.DB
	Validator      *validator.Validator
	UserRepository *repository.UserRepository
}

func NewUserService(logger *logrus.Logger, db *sql.DB, v *validator.Validator, userRepo *repository.UserRepository) *UserService {
	return &UserService{
		Log:            logger,
		DB:             db,
		Validator:      v,
		UserRepository: userRepo,
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]model.UserResponse, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	users, err := s.UserRepository.FindAll(tx)
	if err != nil {
		s.Log.WithError(err).Error("failed to find users")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	response := make([]model.UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, model.UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}

	return response, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.UserResponse, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	u := &entity.User{ID: id}
	if err := s.UserRepository.FindOneByID(tx, u); err != nil {
		s.Log.WithError(err).Error("failed to find user by id")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return &model.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, request *model.CreateUserRequest) (*model.UserResponse, validator.ValidatorErrors, error) {
	fieldErrors := s.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	exists, err := s.UserRepository.ExistsByEmail(tx, request.Email)
	if err != nil {
		s.Log.WithError(err).Error("failed to check user email")
		return nil, nil, err
	}

	if exists {
		return nil, validator.ValidatorErrors{"Error": "Email sudah digunakan"}, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Log.WithError(err).Error("failed to hash password")
		return nil, nil, err
	}

	u := &entity.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hashed),
		Role:     request.Role,
	}

	if err := s.UserRepository.Create(tx, u); err != nil {
		s.Log.WithError(err).Error("failed to create user")
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, err
	}

	return &model.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil, nil
}

func (s *UserService) UpdateUser(ctx context.Context, request *model.UpdateUserRequest) (*model.UserResponse, validator.ValidatorErrors, error) {
	fieldErrors := s.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return nil, fieldErrors, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, nil, err
	}
	defer tx.Rollback()

	u := &entity.User{
		ID:   request.ID,
		Name: request.Name,
		Role: request.Role,
	}

	if err := s.UserRepository.Update(tx, u); err != nil {
		s.Log.WithError(err).Error("failed to update user")
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, err
	}

	return &model.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, nil, nil
}

func (s *UserService) ChangePassword(ctx context.Context, request *model.ChangePasswordRequest) (validator.ValidatorErrors, error) {
	fieldErrors := s.Validator.ValidateStruct(request)
	if fieldErrors != nil {
		return fieldErrors, nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	hashed, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Log.WithError(err).Error("failed to hash password")
		return nil, err
	}

	u := &entity.User{ID: request.ID, Password: string(hashed)}
	if err := s.UserRepository.UpdatePassword(tx, u); err != nil {
		s.Log.WithError(err).Error("failed to update password")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return nil, err
	}

	return nil, nil
}

func (s *UserService) DeleteUser(ctx context.Context, request *model.DeleteUserRequest) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		s.Log.WithError(err).Error("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	u := &entity.User{ID: request.ID}
	if err := s.UserRepository.Delete(tx, u); err != nil {
		s.Log.WithError(err).Error("failed to delete user")
		return err
	}

	if err := tx.Commit(); err != nil {
		s.Log.WithError(err).Error("failed to commit transaction")
		return err
	}

	return nil
}
