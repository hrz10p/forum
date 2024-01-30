package main

import (
	"forum/pkg/models"
	"forum/pkg/services"
	"forum/pkg/utils/logger"
	"net/http"
	"strconv"
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUserFromContext(r)
	if (user == models.User{}) {
		http.Error(w, "cant find a user :(", http.StatusBadRequest)
		return
	}
	postID := r.FormValue("postID")
	postint, err := strconv.Atoi(postID)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		http.Error(w, "postID not correct", http.StatusBadRequest)
		return
	}

	sign := r.FormValue("sign")

	signint, err := strconv.Atoi(sign)
	if err != nil {
		http.Error(w, "sign convert error", http.StatusInternalServerError)
		return
	}

	_, err = h.Service.PostService.GetPostByID(postint)
	if err != nil {
		http.Error(w, "Post not found", http.StatusBadRequest)
		return
	}

	err = h.Service.ReactionService.SubmitReactionForPost(models.Reaction{SubjectID: postint, UID: user.ID, Sign: signint})

	if err != nil {
		switch err {
		case models.SignIsMismatch:
			http.Error(w, "Sign not correct", http.StatusBadRequest)
		default:
			http.Error(w, "Cant react", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
	return
}

func (h *ReactionHandler) ReactComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := getUserFromContext(r)
	if (user == models.User{}) {
		http.Error(w, "cant find a user :(", http.StatusBadRequest)
		return
	}
	postID := r.FormValue("postID")
	postint, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "postID not correct", http.StatusBadRequest)
		return
	}

	commentID := r.FormValue("commentID")
	comint, err := strconv.Atoi(commentID)
	if err != nil {
		http.Error(w, "commentID not correct", http.StatusBadRequest)
		return
	}

	sign := r.FormValue("sign")

	signint, err := strconv.Atoi(sign)
	if err != nil {
		http.Error(w, "sign convert error", http.StatusInternalServerError)
		return
	}

	_, err = h.Service.PostService.GetPostByID(postint)
	if err != nil {
		http.Error(w, "Post not found", http.StatusBadRequest)
		return
	}

	err = h.Service.ReactionService.SubmitReactionForComment(models.Reaction{SubjectID: comint, UID: user.ID, Sign: signint})

	if err != nil {
		switch err {
		case models.SignIsMismatch:
			http.Error(w, "Sign not correct", http.StatusBadRequest)
		default:
			http.Error(w, "Cant react", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
	return
}
