package services

import (
	"database/sql"
	"forum/pkg/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) RegisterUser(user models.User) (models.User, error) {
	NewID := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	user.ID = NewID
	user.Password = string(hash)

	newUser, err := s.createUser(user)
	if err != nil {
		return models.User{}, err
	}
	return newUser, nil
}

func (s *UserService) AuthenticateUser(login, pass string) (models.User, error) {
	user, err := s.getUserByUsername(login)
	if err != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)) != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) createUser(user models.User) (models.User, error) {
	_, err := s.db.Exec("INSERT INTO users (id,username, password, email) VALUES ($1, $2, $3, $4)",
		user.ID,
		user.Username,
		user.Password,
		user.Email)
	if err != nil {
		switch err.Error() {
		case "UNIQUE constraint failed: users.email":
			return models.User{}, models.UniqueConstraintEmail
		case "UNIQUE constraint failed: users.username":
			return models.User{}, models.UniqueConstraintUsername
		default:
			return models.User{}, err
		}
	}

	return user, nil
}

func (s *UserService) GetUserByID(id string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow("SELECT id , username , password , email FROM users WHERE id = $1", id).Scan(&user.ID,
		&user.Username,
		&user.Password,
		&user.Email)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (s *UserService) getUserByUsername(username string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow("SELECT id , username, password , email FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
