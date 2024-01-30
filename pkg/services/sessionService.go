package services

import (
	"database/sql"
	"forum/pkg/models"
	"time"

	"github.com/google/uuid"
)

type SessionService struct {
	db *sql.DB
}

func NewSessionService(db *sql.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) RegisterSession(UID string, exp time.Time) (models.Session, error) {
	existing, err := s.GetSessionByUID(UID)
	if (existing != models.Session{}) {
		if s.DeleteSessionByID(existing.ID) != nil {
			return models.Session{}, err
		}
	}

	ID := uuid.New().String()
	session := models.Session{ID: ID, UID: UID, ExpireTime: exp}

	if s.сreateSession(session) != nil {
		return models.Session{}, nil
	}

	return session, nil
}

func (s *SessionService) сreateSession(session models.Session) error {
	_, err := s.db.Exec("INSERT INTO sessions (id, uid, expireTime) VALUES ($1, $2, $3)", session.ID, session.UID, session.ExpireTime)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) GetSessionByUID(UID string) (models.Session, error) {
	var session models.Session
	err := s.db.QueryRow("SELECT * FROM sessions WHERE uid = $1", UID).Scan(&session.ID, &session.UID, &session.ExpireTime)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (s *SessionService) GetSessionByID(ID string) (models.Session, error) {
	var session models.Session
	err := s.db.QueryRow("SELECT * FROM sessions WHERE id = $1", ID).Scan(&session.ID, &session.UID, &session.ExpireTime)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (s *SessionService) DeleteSessionByID(ID string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
