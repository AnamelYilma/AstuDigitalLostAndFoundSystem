package service

import (
	"errors"
	"fmt"
	"lostfound/internal/model"
	"lostfound/internal/repository"
	"lostfound/pkg/utils"
	"regexp"
	"strings"
	"unicode"
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
	
	// Validate student ID format (must start with "ugr/")
	if !strings.HasPrefix(studentID, "ugr/") {
		return nil, errors.New("student ID must start with 'ugr/' (e.g., ugr/12345/18)")
	}
	
	if phone == "" {
		return nil, errors.New("phone number is required")
	}
	
	// Validate Ethiopian phone number
	if err := validateEthiopianPhone(phone); err != nil {
		return nil, err
	}
	
	// Validate password
	if err := validatePassword(password); err != nil {
		return nil, err
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

func validatePassword(password string) error {
	if len(password) < 7 {
		return errors.New("password must be more than 6 characters")
	}
	
	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		}
	}
	
	if !hasUpper {
		return errors.New("password needs an uppercase letter (A-Z)")
	}
	if !hasLower {
		return errors.New("password needs a lowercase letter (a-z)")
	}
	if !hasNumber {
		return errors.New("password needs a number (0-9)")
	}
	
	return nil
}

func validateEthiopianPhone(phone string) error {
	phone = strings.TrimSpace(phone)
	
	// Pattern 1: 09xxxxxxxx (10 digits starting with 09)
	pattern1 := regexp.MustCompile(`^09\d{8}$`)
	
	// Pattern 2: +2519xxxxxxxx (13 chars: +2519 followed by 8 digits)
	pattern2 := regexp.MustCompile(`^\+2519\d{8}$`)
	
	// Pattern 3: 2519xxxxxxxx (12 digits: 2519 followed by 8 digits)
	pattern3 := regexp.MustCompile(`^2519\d{8}$`)
	
	if pattern1.MatchString(phone) || pattern2.MatchString(phone) || pattern3.MatchString(phone) {
		return nil
	}
	
	// Provide helpful error message
	if len(phone) < 10 {
		return errors.New("phone number is too short. Use 09xxxxxxxx, +2519xxxxxxxx, or 2519xxxxxxxx")
	} else if len(phone) > 13 {
		return errors.New("phone number is too long. Use 09xxxxxxxx, +2519xxxxxxxx, or 2519xxxxxxxx")
	} else if strings.HasPrefix(phone, "09") {
		return errors.New("phone starting with 09 must be exactly 10 digits (09 + 8 digits)")
	} else if strings.HasPrefix(phone, "+251") {
		if !strings.HasPrefix(phone, "+2519") {
			return errors.New("use +2519 followed by 8 digits (e.g., +251912345678)")
		}
		return errors.New("phone with +2519 must be exactly 13 characters (+2519 + 8 digits)")
	} else if strings.HasPrefix(phone, "251") {
		if !strings.HasPrefix(phone, "2519") {
			return errors.New("use 2519 followed by 8 digits (e.g., 251912345678)")
		}
		return errors.New("phone with 2519 must be exactly 12 digits (2519 + 8 digits)")
	}
	
	return errors.New("phone format not recognized. Use 09xxxxxxxx, +2519xxxxxxxx, or 2519xxxxxxxx")
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
