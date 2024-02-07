package main

import (
	"net/http"
	"path/filepath"
)

// InitializeRoutes sets up the application routes
func (app *Application) InitializeRoutes() {
	auth := NewAuthHandler(app.Service)
	post := NewPostHandler(app.Service)
	middle := NewMiddle(app.Service)
	reaction := NewReactionHandler(app.Service)
	comment := NewCommentHandler(app.Service)
	app.Router.Handle("/login", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(http.HandlerFunc(auth.Login))))))
	app.Router.Handle("/register", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(http.HandlerFunc(auth.Registration))))))
	app.Router.Handle("/", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(http.HandlerFunc(post.Index))))))
	app.Router.Handle("/post/", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(http.HandlerFunc(post.Post))))))
	app.Router.Handle("/createPost", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(middle.RequireAuthentication(http.HandlerFunc(post.CreatePost)))))))
	app.Router.Handle("/reactPost", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(middle.RequireAuthentication(http.HandlerFunc(reaction.ReactPost)))))))
	app.Router.Handle("/submitComment", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(middle.RequireAuthentication(http.HandlerFunc(comment.SubmitComment)))))))
	app.Router.Handle("/logout", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(middle.RequireAuthentication(http.HandlerFunc(auth.Logout)))))))
	app.Router.Handle("/reactComment", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(middle.RequireAuthentication(http.HandlerFunc(reaction.ReactComment)))))))

	app.Router.Handle("/catfiltered", middle.Authenticate(middle.LogRequest(middle.RecoverPanic(middle.SecureHeaders(http.HandlerFunc(post.CatFilter))))))

	http.Handle("/ui/style/", http.StripPrefix("/ui/style", http.FileServer(http.Dir("./ui/style/"))))
	app.Logger.Info("routs")
	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/style/")})
	app.Router.Handle("/style", http.NotFoundHandler())
	app.Router.Handle("/style/", http.StripPrefix("/style", fileServer))
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
