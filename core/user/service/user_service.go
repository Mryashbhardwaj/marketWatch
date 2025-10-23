package service

import (
	"context"
	"log/slog"
	"mime/multipart"

	"github.com/Mryashbhardwaj/marketAnalysis/core/user"
)

type SessnionsRepository interface {
	Create(email, password string) (string, error)
	Delete(email, password string) (string, error)
	Update(email, password string) (string, error)
}

type UserRepository interface {
	Get(ctx context.Context, email, password string) (*user.User, error)
	Create(email, password, name, broker string, photoPath *string) (*user.User, error)
	Delete(email, password string) error
}

type UserService struct {
	logger         *slog.Logger
	userRepository UserRepository
}

func NewUserService(logger *slog.Logger, userRepository UserRepository) *UserService {
	return &UserService{
		logger:         logger,
		userRepository: userRepository,
	}
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	s.userRepository.Get(ctx, email, password)
	return "", nil
}

func (s *UserService) Signup(name, email, password, broker string, photoFile multipart.File) error {

	s.userRepository.Create(email, password, name, broker, nil)

	return nil
}
