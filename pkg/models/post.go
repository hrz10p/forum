package models

type Post struct {
	ID      int
	UID     string
	Title   string
	Content string
}

type PostWithCats struct {
	ID      int
	UID     string
	Title   string
	Content string
	Cats    []Category
}
