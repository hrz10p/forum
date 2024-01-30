package services

import "database/sql"

type Service struct {
	UserService     UserService
	PostService     PostService
	CommentService  CommentService
	SessionService  SessionService
	ReactionService ReactionService
}

func NewService(db *sql.DB) *Service {
	return &Service{
		UserService:     *NewUserService(db),
		PostService:     *NewPostService(db),
		ReactionService: *NewReactionService(db),
		CommentService:  *NewCommentService(db),
		SessionService:  *NewSessionService(db),
	}
}
