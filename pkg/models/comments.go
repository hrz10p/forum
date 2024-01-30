package models

type Comment struct {
	ID      int
	UID     string
	PostID  int
	Content string
}
