package services

import (
	"database/sql"
	"forum/pkg/models"
)

type CommentService struct {
	db *sql.DB
}

func NewCommentService(db *sql.DB) *CommentService {
	return &CommentService{db: db}
}

func (s *CommentService) SubmitCommentForPost(comment models.Comment) error {
	_, err := s.db.Exec("INSERT INTO comments (uid, post_id, content) VALUES ($1, $2, $3)",
		comment.UID, comment.PostID, comment.Content)

	return err
}

func (s *CommentService) DeleteComment(ID int) error {
	_, err := s.db.Exec("DELETE FROM comments WHERE id = $1", ID)

	return err
}

func (s *CommentService) GetCommentsByPostID(postID int) ([]models.Comment, error) {
	rows, err := s.db.Query("SELECT * FROM comments WHERE post_id = $1", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment

	for rows.Next() {
		comment := models.Comment{}

		err := rows.Scan(&comment.ID,
			&comment.UID,
			&comment.PostID,
			&comment.Content,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
