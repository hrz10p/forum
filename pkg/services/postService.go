package services

import (
	"database/sql"
	"fmt"
	"forum/pkg/models"
	"strconv"
)

type PostService struct {
	db *sql.DB
}

func NewPostService(db *sql.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) GetAllPosts() ([]models.PostWithCats, error) {
	rows, err := s.db.Query("SELECT * FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithCats

	for rows.Next() {
		post := models.Post{}

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UID,
		)
		if err != nil {
			return nil, err
		}

		cats, err := s.getCatsForPost(post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, models.PostWithCats{
			ID:      post.ID,
			UID:     post.UID,
			Title:   post.Title,
			Content: post.Content,
			Cats:    cats,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) GetReactedPosts(UID string) ([]models.PostWithCats, error) {
	rows, err := s.db.Query("SELECT p.id, p.title, p.content, p.uid FROM posts p JOIN post_reacts a ON p.id = a.pid WHERE a.uid = $1 AND a.sign IN (1, -1)", UID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, models.NotFoundAnything
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var posts []models.PostWithCats

	for rows.Next() {
		post := models.Post{}

		err := rows.Scan(&post.ID,
			&post.Title,
			&post.Content,
			&post.UID,
		)
		cats, err := s.getCatsForPost(post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, models.PostWithCats{
			ID:      post.ID,
			UID:     post.UID,
			Title:   post.Title,
			Content: post.Content,
			Cats:    cats,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) CreatePost(p models.Post, catIDS []int) error {
	if len(catIDS) < 1 {
		return models.NoCatsSelected
	}

	result, err := s.db.Exec("INSERT INTO posts (title , content , UID) VALUES ($1 , $2 , $3)", p.Title, p.Content, p.UID)
	if err != nil {
		return err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if err := s.insertCatsForPost(int(newID), catIDS); err != nil {
		return err
	}

	return nil
}

func (s *PostService) GetPostByID(ID int) (models.PostWithCats, error) {
	if ID < 1 {
		return models.PostWithCats{}, models.ValueMismatch
	}

	post := models.Post{}

	err := s.db.QueryRow("SELECT * FROM posts WHERE id = $1", ID).Scan(&post.ID,
		&post.Title,
		&post.Content,
		&post.UID,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models.PostWithCats{}, models.NotFoundAnything
		default:
			return models.PostWithCats{}, err
		}
	}

	cats, err := s.getCatsForPost(post.ID)
	if err != nil {
		return models.PostWithCats{}, err
	}

	return models.PostWithCats{
		ID:      post.ID,
		UID:     post.UID,
		Title:   post.Title,
		Content: post.Content,
		Cats:    cats,
	}, nil
}

func (s *PostService) insertCatsForPost(postID int, catIDS []int) error {
	stmt, err := s.db.Prepare("INSERT INTO post_cats (post_id, category_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, cat := range catIDS {
		_, err := stmt.Exec(postID, cat)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostService) GetPostsByCats(catIDS []int) ([]models.PostWithCats, error) {
	query := `
		SELECT DISTINCT p.id, p.uid, p.title, p.content
		FROM posts p
		JOIN post_cats pc ON p.id = pc.post_id
		WHERE pc.category_id IN (%s);
	`

	catIDsStr := ""
	for i, id := range catIDS {
		if i > 0 {
			catIDsStr += ","
		}
		catIDsStr += strconv.Itoa(id)
	}

	makeQuery := fmt.Sprintf(query, catIDsStr)

	rows, err := s.db.Query(makeQuery)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, models.NotFoundAnything
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var posts []models.PostWithCats

	for rows.Next() {
		post := models.Post{}

		err := rows.Scan(&post.ID, &post.UID, &post.Title, &post.Content)
		if err != nil {
			return nil, err
		}

		cats, err := s.getCatsForPost(post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, models.PostWithCats{
			ID:      post.ID,
			UID:     post.UID,
			Title:   post.Title,
			Content: post.Content,
			Cats:    cats,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) getCatsForPost(postID int) ([]models.Category, error) {
	rows, err := s.db.Query(`
        SELECT c.id, c.name
        FROM categories c
        JOIN post_cats pc ON c.id = pc.category_id
        WHERE pc.post_id = $1
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		category := models.Category{}

		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *PostService) GetCats() ([]models.Category, error) {
	rows, err := s.db.Query(`
        SELECT c.id, c.name
        FROM categories c`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		category := models.Category{}

		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *PostService) GetPostsByUID(UID string) ([]models.PostWithCats, error) {
	rows, err := s.db.Query("SELECT * FROM posts WHERE uid = $1", UID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostWithCats

	for rows.Next() {
		post := models.Post{}

		err := rows.Scan(&post.ID,
			&post.Title,
			&post.Content,
			&post.UID,
		)
		if err != nil {
			return nil, err
		}

		cats, err := s.getCatsForPost(post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, models.PostWithCats{
			ID:      post.ID,
			UID:     post.UID,
			Title:   post.Title,
			Content: post.Content,
			Cats:    cats,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *PostService) DeletePost(ID int) error {
	_, err := s.db.Exec("DELETE FROM posts WHERE id = $1", ID)
	if err != nil {
		return err
	}

	return nil
}
