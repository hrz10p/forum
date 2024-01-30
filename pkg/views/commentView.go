package views

type CommentView struct {
	ID            int
	Author        string
	Content       string
	LikesCount    int
	DislikesCount int
	IsLiked       bool
	IsDisliked    bool
}
