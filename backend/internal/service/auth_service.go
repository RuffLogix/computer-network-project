package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rufflogix/computer-network-project/internal/entity"
	"github.com/rufflogix/computer-network-project/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(username, password, name, email string) (*entity.User, string, error)
	Login(username, password string) (*entity.User, string, error)
	CreateGuestUser(name string) (*entity.User, string, error)
	ValidateToken(tokenString string) (*entity.User, error)
	GetUserByNumericID(id int64) (*entity.User, error)
}

type implAuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &implAuthService{
		userRepo: userRepo,
	}
}

func (s *implAuthService) Register(username, password, name, email string) (*entity.User, string, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetUserByUsername(username)
	if existingUser != nil {
		return nil, "", errors.New("username already exists")
	}

	existingEmail, _ := s.userRepo.GetUserByEmail(email)
	if existingEmail != nil {
		return nil, "", errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Create user
	user := &entity.User{
		Username:  username,
		Password:  string(hashedPassword),
		Name:      name,
		Email:     email,
		IsGuest:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *implAuthService) Login(username, password string) (*entity.User, string, error) {
	// Find user
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("invalid username or password")
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *implAuthService) CreateGuestUser(name string) (*entity.User, string, error) {
	// Create guest user
	user := &entity.User{
		Username:  "guest_" + primitive.NewObjectID().Hex(),
		Name:      name,
		IsGuest:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *implAuthService) ValidateToken(tokenString string) (*entity.User, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user ID")
		}

		user, err := s.userRepo.GetUserByID(userID)
		if err != nil {
			return nil, errors.New("user not found")
		}

		return user, nil
	}

	return nil, errors.New("invalid token")
}

func (s *implAuthService) generateToken(user *entity.User) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"is_guest": user.IsGuest,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func (s *implAuthService) GetUserByNumericID(id int64) (*entity.User, error) {
	return s.userRepo.GetUserByNumericID(id)
}
