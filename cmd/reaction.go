package main

import (
	"net/http"
	"strconv"

	"forum/pkg/models"
	"forum/pkg/services"
	"forum/pkg/utils/logger"
)

type ReactionHandler struct {
	Service *services.Service
}

func NewReactionHandler(Service *services.Service) *ReactionHandler {
	return &ReactionHandler{
		Service: Service,
	}
}

func (h *ReactionHandler) ReactPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		ErrorPage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUserFromContext(r)
	if (user == models.User{}) {
		ErrorPage(w, "cant find a user :(", http.StatusBadRequest)
		return
	}
	postID := r.FormValue("postID")
	postint, err := strconv.Atoi(postID)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		ErrorPage(w, "postID not correct", http.StatusBadRequest)
		return
	}

	sign := r.FormValue("sign")

	signint, err := strconv.Atoi(sign)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		ErrorPage(w, "sign convert error", http.StatusInternalServerError)
		return
	}

	_, err = h.Service.PostService.GetPostByID(postint)
	if err != nil {
		ErrorPage(w, "Post not found", http.StatusBadRequest)
		return
	}

	err = h.Service.ReactionService.SubmitReactionForPost(models.Reaction{SubjectID: postint, UID: user.ID, Sign: signint})

	if err != nil {
		switch err {
		case models.SignIsMismatch:
			ErrorPage(w, "Sign not correct", http.StatusBadRequest)
		default:
			ErrorPage(w, "Cant react", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
	return
}

func (h *ReactionHandler) ReactComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		ErrorPage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUserFromContext(r)
	if (user == models.User{}) {
		ErrorPage(w, "cant find a user :(", http.StatusBadRequest)
		return
	}
	postID := r.FormValue("postID")
	postint, err := strconv.Atoi(postID)
	if err != nil {
		ErrorPage(w, "postID not correct", http.StatusBadRequest)
		return
	}

	commentID := r.FormValue("commentID")
	comint, err := strconv.Atoi(commentID)
	if err != nil {
		ErrorPage(w, "commentID not correct", http.StatusBadRequest)
		return
	}

	sign := r.FormValue("sign")

	signint, err := strconv.Atoi(sign)
	if err != nil {
		ErrorPage(w, "sign convert error", http.StatusInternalServerError)
		return
	}

	_, err = h.Service.PostService.GetPostByID(postint)
	if err != nil {
		ErrorPage(w, "Post not found", http.StatusBadRequest)
		return
	}

	err = h.Service.ReactionService.SubmitReactionForComment(models.Reaction{SubjectID: comint, UID: user.ID, Sign: signint})

	if err != nil {
		switch err {
		case models.SignIsMismatch:
			ErrorPage(w, "Sign not correct", http.StatusBadRequest)
		default:
			ErrorPage(w, "Cant react", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
	return
}
