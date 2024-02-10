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
	"forum/pkg/utils/validators"
	"forum/pkg/views"
)

type PostHanlder struct {
	Service *services.Service
}

type page struct {
	Auth     bool
	Username string
	Cats     []views.CategoryView
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

func (p *PostHanlder) catToViewConverter(cats []models.Category) []views.CategoryView {
	var viewss []views.CategoryView
	for _, val := range cats {
		viewss = append(viewss, views.CategoryView{ID: val.ID, Name: val.Name, Checked: false})
	}
	return viewss
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
			logger.GetLogger().Error(err.Error())
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
		ErrorPage(w, "Not found", 404)
		logger.GetLogger().Info("NOTF" + r.URL.Path)
		return
	}
	file := "./ui/templates/index.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		ErrorPage(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)

	posts, err := p.Service.PostService.GetAllPosts()
	if err != nil {
		ErrorPage(w, "Cant fecth posts", http.StatusInternalServerError)
		return
	}

	views, err := p.converterPOSTS(posts)
	if err != nil {
		ErrorPage(w, "Cant load views", http.StatusInternalServerError)
		return
	}

	cats, err := p.Service.PostService.GetCats()
	if err != nil {
		ErrorPage(w, "Cant fecth cats", http.StatusInternalServerError)
		return
	}

	catViews := p.catToViewConverter(cats)

	data := page{
		Posts: views,
		Cats:  catViews,
	}

	if (user != models.User{}) {
		data.Auth = true
		data.Username = user.Username
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		ErrorPage(w, "Error executing template", 500)
		return
	}
}

func (p *PostHanlder) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			ErrorPage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUserFromContext(r)
		if (user == models.User{}) {
			ErrorPage(w, "cant find a user :(", http.StatusInternalServerError)
			return
		}
		content := r.FormValue("content")
		title := r.FormValue("title")
		cats := r.Form["cats"]
		if cats == nil {
			ErrorPage(w, "No cats selected", http.StatusBadRequest)
			return
		}
		catIds, err := p.stringsToInts(cats)

		if validators.LengthRangeValidate(content, 5, 300) != nil {
			ErrorPage(w, "Content is to short or too long", 400)
			return
		}

		if validators.LengthRangeValidate(title, 5, 15) != nil {
			ErrorPage(w, "Title is to short or too long", 400)
			return
		}

		if p.Service.PostService.CreatePost(models.Post{
			UID:     user.ID,
			Title:   title,
			Content: content,
		}, catIds) != nil {
			ErrorPage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else if r.Method == http.MethodGet {
		file := "./ui/templates/postCreate.html"
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			ErrorPage(w, "Error parsing templates", 500)
			return
		}
		cats, err := p.Service.PostService.GetCats()
		if err != nil {
			ErrorPage(w, "Error parsing cats", 500)
			return
		}
		err = tmpl.Execute(w, cats)
		if err != nil {
			logger.GetLogger().Warn(err.Error())
			ErrorPage(w, "Error executing template", 500)
			return
		}
		return
	} else {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (p *PostHanlder) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file := "./ui/templates/post.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		ErrorPage(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)

	if !strings.HasPrefix(r.URL.Path, "/post/") {
		ErrorPage(w, "Page not found", http.StatusNotFound)
		return
	}

	pathID := r.URL.Path[len("/post/"):]
	postID, err := strconv.Atoi(pathID)
	if err != nil {
		ErrorPage(w, "Page not found", http.StatusNotFound)
		return
	}
	post, err := p.Service.PostService.GetPostByID(postID)
	if err != nil {
		switch err {
		case models.NotFoundAnything:
			ErrorPage(w, "Not found post", http.StatusNotFound)
			break
		default:
			logger.GetLogger().Error(err.Error())
			ErrorPage(w, "Post load problem", http.StatusInternalServerError)
		}
		return
	}

	postview, err := p.convertPostToView(post)
	if err != nil {
		ErrorPage(w, "Error converting post", http.StatusInternalServerError)
		return
	}

	comments, err := p.Service.CommentService.GetCommentsByPostID(postID)
	if err != nil {
		ErrorPage(w, "Cant load comments", http.StatusInternalServerError)
		return
	}

	comviews, err := p.convertCommentToView(comments)
	if err != nil {
		ErrorPage(w, "Error converting comments", http.StatusInternalServerError)
		return
	}

	Likes, Dislikes, err := p.Service.ReactionService.GetReactionCountsForPost(postID)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		ErrorPage(w, "Error getting reaction counts", http.StatusInternalServerError)
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
			ErrorPage(w, "Cant load reaction for post", http.StatusInternalServerError)
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
				ErrorPage(w, "Cant load reactions for comments", http.StatusInternalServerError)
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
		ErrorPage(w, "Error executing template", 500)
		return
	}
}

func (p *PostHanlder) CatFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file := "./ui/templates/index.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		ErrorPage(w, "Error parsing templates", 500)
		return
	}

	err = r.ParseForm()
	if err != nil {
		ErrorPage(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	catsFORM := r.Form["category"]

	fmt.Println(catsFORM)
	catsINTS, err := p.stringsToInts(catsFORM)

	user := getUserFromContext(r)
	if err != nil {
		ErrorPage(w, "Cant convert cats to int", http.StatusInternalServerError)
		return

	}

	posts, err := p.Service.PostService.GetPostsByCats(catsINTS)
	if err != nil {
		ErrorPage(w, "Cant fecth posts", http.StatusInternalServerError)
		return
	}

	views, err := p.converterPOSTS(posts)
	if err != nil {
		ErrorPage(w, "Cant load views", http.StatusInternalServerError)
		return
	}

	cats, err := p.Service.PostService.GetCats()
	if err != nil {
		ErrorPage(w, "Cant fecth cats", http.StatusInternalServerError)
		return
	}

	catViews := p.catToViewConverter(cats)

	for i, val := range catViews {
		for _, val2 := range catsINTS {
			if val.ID == val2 {
				catViews[i].Checked = true
			}
		}
	}

	data := page{
		Posts: views,
		Cats:  catViews,
	}

	if (user != models.User{}) {
		data.Auth = true
		data.Username = user.Username
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		ErrorPage(w, "Error executing template", 500)
		return
	}
}

func (p *PostHanlder) Created(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file := "./ui/templates/index.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		ErrorPage(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)
	if err != nil {
		ErrorPage(w, "Cant convert cats to int", http.StatusInternalServerError)
		return
	}

	posts, err := p.Service.PostService.GetAllPosts()

	var postsCreatedByUser []models.PostWithCats

	for _, val := range posts {
		if val.UID == user.ID {
			postsCreatedByUser = append(postsCreatedByUser, val)
		}
	}
	if err != nil {
		ErrorPage(w, "Cant fecth posts", http.StatusInternalServerError)
		return
	}

	views, err := p.converterPOSTS(postsCreatedByUser)
	if err != nil {
		ErrorPage(w, "Cant load views", http.StatusInternalServerError)
		return
	}

	cats, err := p.Service.PostService.GetCats()
	if err != nil {
		ErrorPage(w, "Cant fecth cats", http.StatusInternalServerError)
		return
	}

	catViews := p.catToViewConverter(cats)

	data := page{
		Posts: views,
		Cats:  catViews,
	}

	if (user != models.User{}) {
		data.Auth = true
		data.Username = user.Username
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		ErrorPage(w, "Error executing template", 500)
		return
	}
}

func (p *PostHanlder) Reacted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	file := "./ui/templates/index.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		ErrorPage(w, "Error parsing templates", 500)
		return
	}

	user := getUserFromContext(r)
	if err != nil {
		ErrorPage(w, "Cant convert cats to int", http.StatusInternalServerError)
		return
	}

	posts, err := p.Service.PostService.GetReactedPosts(user.ID)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		ErrorPage(w, "Cant fecth posts", http.StatusInternalServerError)
		return
	}

	views, err := p.converterPOSTS(posts)
	if err != nil {
		ErrorPage(w, "Cant load views", http.StatusInternalServerError)
		return
	}

	cats, err := p.Service.PostService.GetCats()
	if err != nil {
		ErrorPage(w, "Cant fecth cats", http.StatusInternalServerError)
		return
	}

	catViews := p.catToViewConverter(cats)

	data := page{
		Posts: views,
		Cats:  catViews,
	}

	if (user != models.User{}) {
		data.Auth = true
		data.Username = user.Username
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.GetLogger().Warn(err.Error())
		ErrorPage(w, "Error executing template", 500)
		return
	}
}
