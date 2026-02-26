package service

import (
    "errors"
    "lostfound/internal/model"
    "lostfound/internal/repository"
    "lostfound/pkg/utils"
    // "github.com/google/uuid" 
)

type AuthService struct {
    userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
    return &AuthService{
        userRepo: repository.NewUserRepository(),
    }
}

func (s *AuthService) Register(name, email, password string) (*model.User, error) {
    existingUser, _ := s.userRepo.FindByEmail(email)
    if existingUser != nil && existingUser.ID > 0 {
        return nil, errors.New("email already registered")
    }
    
    hashedPassword, err := utils.HashPassword(password)
    if err != nil {
        return nil, err
    }
    
    user := &model.User{
        Name:     name,
        Email:    email,
        Password: hashedPassword,
        Role:     "student",
    }
    
    err = s.userRepo.Create(user)
    return user, err
}

func (s *AuthService) Login(email, password string) (*model.User, error) {
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return nil, errors.New("invalid email or password")
    }
    
    if !utils.CheckPasswordHash(password, user.Password) {
        return nil, errors.New("invalid email or password")
    }
    
    return user, nil
}