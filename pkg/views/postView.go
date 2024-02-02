package views

import "forum/pkg/models"

type PostView struct {
	AuthorName string
	Title      string
	Content    string
	Cats       []models.Category
	Id         int
}
