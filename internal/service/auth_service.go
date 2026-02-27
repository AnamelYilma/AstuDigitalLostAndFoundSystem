package service

import (
	"errors"
	"fmt"
	"lostfound/internal/model"
	"lostfound/internal/repository"
	"lostfound/pkg/utils"
	"strings"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *AuthService) Register(name, studentID, phone, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	studentID = strings.ToLower(strings.TrimSpace(studentID))
	phone = strings.TrimSpace(phone)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if studentID == "" {
		return nil, errors.New("student ID is required")
	}
	if phone == "" {
		return nil, errors.New("phone number is required")
	}
	if len(strings.TrimSpace(password)) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	existingUser, _ := s.userRepo.FindByStudentID(studentID)
	if existingUser != nil && existingUser.ID > 0 {
		return nil, errors.New("student ID already registered")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:      name,
		StudentID: studentID,
		Phone:     phone,
		Email:     fmt.Sprintf("%s@astu.local", strings.ReplaceAll(strings.ToLower(studentID), "/", "_")),
		Password:  hashedPassword,
		Role:      "student",
	}

	err = s.userRepo.Create(user)
	return user, err
}

func (s *AuthService) Login(studentID, password string) (*model.User, error) {
	studentID = strings.ToLower(strings.TrimSpace(studentID))
	user, err := s.userRepo.FindByStudentID(studentID)
	if err != nil {
		return nil, errors.New("invalid ID or password")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid ID or password")
	}

	return user, nil
}
