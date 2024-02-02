package services

import (
	"database/sql"
	"forum/pkg/models"
)

type ReactionService struct {
	db *sql.DB
}

func NewReactionService(db *sql.DB) *ReactionService {
	return &ReactionService{db: db}
}

func (s *ReactionService) SubmitReactionForPost(reaction models.Reaction) error {
	if reaction.Sign != 1 && reaction.Sign != -1 {
		return models.SignIsMismatch
	}
	existingSign, err := s.GetReactionSignForPost(reaction.UID, reaction.SubjectID)
	if err != nil {
		return err
	}

	if existingSign == reaction.Sign {
		return s.DeleteReactionForPost(reaction.UID, reaction.SubjectID)
	}

	if existingSign != 0 {
		return s.SwapReactionForPost(reaction.UID, reaction.SubjectID, reaction.Sign)
	}

	return s.InsertReactionForPost(reaction)
}

func (s *ReactionService) SubmitReactionForComment(reaction models.Reaction) error {
	if reaction.Sign != 1 && reaction.Sign != -1 {
		return models.SignIsMismatch
	}
	existingSign, err := s.GetReactionSignForComment(reaction.UID, reaction.SubjectID)
	if err != nil {
		return err
	}

	if existingSign == reaction.Sign {
		return s.DeleteReactionForComment(reaction.UID, reaction.SubjectID)
	}

	if existingSign != 0 {
		return s.SwapReactionForComment(reaction.UID, reaction.SubjectID, reaction.Sign)
	}

	return s.InsertReactionForComment(reaction)
}

func (s *ReactionService) GetReactionCountsForPost(postId int) (int, int, error) {
	var likes, dislikes int
	err := s.db.QueryRow("SELECT COUNT(*) FROM posts_reactions WHERE post_id = ? AND sign = 1", postId).Scan(&likes)
	if err != nil {
		return 0, 0, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM posts_reactions WHERE post_id = ? AND sign = -1", postId).Scan(&dislikes)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}

func (s *ReactionService) GetReactionCountsForComment(comId int) (int, int, error) {
	var likes, dislikes int
	err := s.db.QueryRow("SELECT COUNT(*) FROM comments_reactions WHERE comment_id = ? AND sign = 1", comId).Scan(&likes)
	if err != nil {
		return 0, 0, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM comments_reactions WHERE comment_id = ? AND sign = -1", comId).Scan(&dislikes)
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}

func (s *ReactionService) GetReactionSignForPost(uid string, postID int) (int, error) {
	var sign int
	err := s.db.QueryRow("SELECT sign FROM posts_reactions WHERE user_id = $1 AND post_id = $2", uid, postID).Scan(&sign)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return sign, err
}

func (s *ReactionService) GetReactionSignForComment(uid string, commentID int) (int, error) {
	var sign int
	err := s.db.QueryRow("SELECT sign FROM comments_reactions WHERE user_id = $1 AND comment_id = $2", uid, commentID).Scan(&sign)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return sign, err
}

func (s *ReactionService) InsertReactionForPost(reaction models.Reaction) error {
	_, err := s.db.Exec("INSERT INTO posts_reactions (user_id, post_id, sign) VALUES ($1, $2, $3)", reaction.UID, reaction.SubjectID, reaction.Sign)
	return err
}

func (s *ReactionService) InsertReactionForComment(reaction models.Reaction) error {
	_, err := s.db.Exec("INSERT INTO comments_reactions (user_id, comment_id, sign) VALUES ($1, $2, $3)", reaction.UID, reaction.SubjectID, reaction.Sign)
	return err
}

func (s *ReactionService) DeleteReactionForPost(uid string, postID int) error {
	_, err := s.db.Exec("DELETE FROM posts_reactions WHERE user_id = $1 AND post_id = $2", uid, postID)
	return err
}

func (s *ReactionService) DeleteReactionForComment(uid string, commentID int) error {
	_, err := s.db.Exec("DELETE FROM comments_reactions WHERE user_id = $1 AND comment_id = $2", uid, commentID)
	return err
}

func (s *ReactionService) SwapReactionForPost(uid string, postID int, newSign int) error {
	_, err := s.db.Exec("UPDATE posts_reactions SET sign = $1 WHERE user_id = $2 AND post_id = $3", newSign, uid, postID)
	return err
}

func (s *ReactionService) SwapReactionForComment(uid string, commentID int, newSign int) error {
	_, err := s.db.Exec("UPDATE comments_reactions SET sign = $1 WHERE user_id = $2 AND comment_id = $3", newSign, uid, commentID)
	return err
}
