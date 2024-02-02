package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"forum/pkg/models"
	"forum/pkg/services"
	"forum/pkg/utils/logger"
	"forum/pkg/views"
)

type PostHanlder struct {
	Service *services.Service
}

type page struct {
	Auth     bool
	Username string
	Cats     []models.Category
	Posts    []views.PostView
}

type showPost struct {
	Auth          bool
	Post          views.PostView
	Comments      []views.CommentView
	LikesCount    int
	DislikesCount int
	IsLiked       bool
	IsDisliked    bool
}

func (p *PostHanlder) stringsToInts(str []string) ([]int, error) {
	var ids []int
	for _, val := range str {
		num, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		ids = append(ids, num)
	}
	return ids, nil
}

func (p *PostHanlder) convertPostToView(post models.PostWithCats) (views.PostView, error) {
	user, err := p.Service.UserService.GetUserByID(post.UID)
	if err != nil {
		return views.PostView{}, err
	}

	return views.PostView{
		Id:         post.ID,
		AuthorName: user.Username,
		Content:    post.Content,
		Title:      post.Title,
		Cats:       post.Cats,
	}, nil
}

func (p *PostHanlder) convertCommentToView(comments []models.Comment) ([]views.CommentView, error) {
	var v []views.CommentView
	for _, val := range comments {
		user, err := p.Service.UserService.GetUserByID(val.UID)
		if err != nil {
			return nil, err
		}

		likes, dislikes, err := p.Service.ReactionService.GetReactionCountsForComment(val.ID)
		if err != nil {
			return nil, err
		}
		v = append(v, views.CommentView{Author: user.Username, Content: val.Content, ID: val.ID, IsLiked: false, IsDisliked: false, LikesCount: likes, DislikesCount: dislikes})
	}

	return v, nil
}

func (p *PostHanlder) converterPOSTS(posts []models.PostWithCats) ([]views.PostView, error) {
	var views []views.PostView
	for _, val := range posts {
		view, err := p.convertPostToView(val)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func NewPostHandler(Service *services.Service) *PostHanlder {
	return &PostHanlder{
		Service: Service,
	}
}

func (p *PostHanlder) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	file := "./ui/templates/index.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		http.Error(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)

	posts, err := p.Service.PostService.GetAllPosts()
	if err != nil {
		http.Error(w, "Cant fecth posts", http.StatusInternalServerError)
		return
	}

	views, err := p.converterPOSTS(posts)
	if err != nil {
		http.Error(w, "Cant load views", http.StatusInternalServerError)
		return
	}

	cats, err := p.Service.PostService.GetCats()
	if err != nil {
		http.Error(w, "Cant fecth cats", http.StatusInternalServerError)
		return
	}

	data := page{
		Posts: views,
		Cats:  cats,
	}

	if (user != models.User{}) {
		data.Auth = true
		data.Username = user.Username
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		http.Error(w, "Error executing template", 500)
		return
	}
}

func (p *PostHanlder) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUserFromContext(r)
		if (user == models.User{}) {
			http.Error(w, "cant find a user :(", http.StatusInternalServerError)
			return
		}
		content := r.FormValue("content")
		title := r.FormValue("title")
		cats := r.Form["cats"]
		fmt.Println(content, cats)
		catIds, err := p.stringsToInts(cats)

		if p.Service.PostService.CreatePost(models.Post{
			UID:     user.ID,
			Title:   title,
			Content: content,
		}, catIds) != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else if r.Method == http.MethodGet {
		file := "./ui/templates/postCreate.html"
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			http.Error(w, "Error parsing templates", 500)
			return
		}
		cats, err := p.Service.PostService.GetCats()
		if err != nil {
			http.Error(w, "Error parsing cats", 500)
			return
		}
		err = tmpl.Execute(w, cats)
		if err != nil {
			logger.GetLogger().Warn(err.Error())
			http.Error(w, "Error executing template", 500)
			return
		}
		return
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (p *PostHanlder) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file := "./ui/templates/post.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		http.Error(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)

	if !strings.HasPrefix(r.URL.Path, "/post/") {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	pathID := r.URL.Path[len("/post/"):]
	postID, err := strconv.Atoi(pathID)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	post, err := p.Service.PostService.GetPostByID(postID)
	if err != nil {
		switch err {
		case models.NotFoundAnything:
			http.Error(w, "Not found post", http.StatusNotFound)
			break
		default:
			logger.GetLogger().Error(err.Error())
			http.Error(w, "Post load problem", http.StatusInternalServerError)
		}
		return
	}

	postview, err := p.convertPostToView(post)
	if err != nil {
		http.Error(w, "Error converting post", http.StatusInternalServerError)
		return
	}

	comments, err := p.Service.CommentService.GetCommentsByPostID(postID)
	if err != nil {
		http.Error(w, "Cant load comments", http.StatusInternalServerError)
		return
	}

	comviews, err := p.convertCommentToView(comments)
	if err != nil {
		http.Error(w, "Error converting comments", http.StatusInternalServerError)
		return
	}

	Likes, Dislikes, err := p.Service.ReactionService.GetReactionCountsForPost(postID)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		http.Error(w, "Error getting reaction counts", http.StatusInternalServerError)
		return
	}

	data := showPost{
		Post:          postview,
		Comments:      comviews,
		LikesCount:    Likes,
		DislikesCount: Dislikes,
		IsLiked:       false,
		IsDisliked:    false,
	}
	if (user != models.User{}) {
		data.Auth = true
		sign, err := p.Service.ReactionService.GetReactionSignForPost(user.ID, postID)
		if err != nil {
			http.Error(w, "Cant load reaction for post", http.StatusInternalServerError)
			return
		}

		switch sign {
		case 1:
			data.IsLiked = true
			break
		case -1:
			data.IsDisliked = true
			break
		}

		for i, val := range data.Comments {
			r, err := p.Service.ReactionService.GetReactionSignForComment(user.ID, val.ID)
			if err != nil {
				http.Error(w, "Cant load reactions for comments", http.StatusInternalServerError)
				return
			}
			switch r {
			case 1:
				data.Comments[i].IsLiked = true
				break
			case -1:
				data.Comments[i].IsLiked = true
				break
			}

		}
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		http.Error(w, "Error executing template", 500)
		return
	}
}
