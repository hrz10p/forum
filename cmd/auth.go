package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"forum/pkg/models"
	"forum/pkg/services"
	"forum/pkg/utils/cookies"
	"forum/pkg/utils/logger"
	"forum/pkg/utils/validators"
)

type ErrorMessages struct {
	LoginError                string
	EmailError                string
	PasswordError             string
	PasswordConfirmationError string
}

type AuthHanlder struct {
	Service *services.Service
}

func NewAuthHandler(Service *services.Service) *AuthHanlder {
	return &AuthHanlder{
		Service: Service,
	}
}

func (a *AuthHanlder) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			ErrorPage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		login := r.FormValue("username")
		pass := r.FormValue("password")

		if validators.LengthRangeValidate(login, 2, 10) != nil || validators.PasswordValidate(pass) != nil {
			file := "./ui/templates/login.html"
			tmpl, err := template.ParseFiles(file)
			if err != nil {
				ErrorPage(w, "Error parsing templates", 500)
				return
			}
			w.WriteHeader(400)
			err = tmpl.Execute(w, "Invalid username or password")
			if err != nil {
				fmt.Print(err)
				ErrorPage(w, "Error executing template", 500)
				return
			}
			return
		}

		user, err := a.Service.UserService.AuthenticateUser(login, pass)
		if err != nil {
			switch err {
			case models.ErrInvalidCredentials:
				file := "./ui/templates/login.html"
				tmpl, err := template.ParseFiles(file)
				if err != nil {
					ErrorPage(w, "Error parsing templates", 500)
					return
				}
				w.WriteHeader(400)
				err = tmpl.Execute(w, "Invalid username or password")
				if err != nil {
					fmt.Print(err)
					ErrorPage(w, "Error executing template", 500)
					return
				}
				return

			default:
				ErrorPage(w, "UNKNOWN", http.StatusInternalServerError)
				return
			}
		}

		times := time.Now().Add(time.Hour)

		session, err := a.Service.SessionService.RegisterSession(user.ID, times)
		if err != nil {
			logger.GetLogger().Warn(err.Error())
			ErrorPage(w, "ERROR CREATING SESSION", http.StatusInternalServerError)
			return
		}

		cookies.SetCookie(w, session.ID, times) // RETURN TO THIS
		fmt.Println(session)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else if r.Method == http.MethodGet {
		file := "./ui/templates/login.html"
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			ErrorPage(w, "Error parsing templates", 500)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			fmt.Print(err)
			ErrorPage(w, "Error executing template", 500)
			return
		}
	} else {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *AuthHanlder) Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			ErrorPage(w, err.Error(), http.StatusInternalServerError)
			return
		}

		login := r.FormValue("username")
		email := r.FormValue("email")
		pass := r.FormValue("password")
		passConf := r.FormValue("passwordConf")
		fmt.Println(login, email, pass, passConf)
		var errorMessages ErrorMessages

		if validators.LengthRangeValidate(login, 2, 10) != nil {
			errorMessages.LoginError = "Username must be between 2 and 10 characters"
		}

		if validators.PasswordValidate(pass) != nil {
			errorMessages.PasswordError = "Password must have at least 5 characters and no more than 20 characters"
		}

		if validators.EmailValidate(email) != nil {
			errorMessages.EmailError = "Invalid email format"
		}

		if pass != passConf {
			errorMessages.PasswordError = "Passwords doesn't match"
		}

		if errorMessages.LoginError != "" || errorMessages.PasswordError != "" || errorMessages.EmailError != "" || errorMessages.PasswordConfirmationError != "" {
			file := "./ui/templates/reg.html"
			tmpl, err := template.ParseFiles(file)
			if err != nil {
				ErrorPage(w, "Error parsing templates", 500)
				return
			}
			w.WriteHeader(400)
			err = tmpl.Execute(w, errorMessages)
			if err != nil {
				fmt.Print(err)
				ErrorPage(w, "Error executing template", 500)
				return
			}
			return
		}

		_, err = a.Service.UserService.RegisterUser(models.User{Username: login, Email: email, Password: pass})

		if err != nil {
			switch err {
			case models.UniqueConstraintUsername:
				errorMessages.LoginError = "Username exists"
				break
			case models.UniqueConstraintEmail:
				errorMessages.EmailError = "Email exists"
				break
			default:
				ErrorPage(w, "UNKNOWN ERROR", http.StatusInternalServerError)
				logger.GetLogger().Error(err.Error())
				return

			}
			file := "./ui/templates/reg.html"
			tmpl, err := template.ParseFiles(file)
			if err != nil {
				ErrorPage(w, "Error parsing templates", 500)
				return
			}
			w.WriteHeader(400)
			err = tmpl.Execute(w, errorMessages)
			if err != nil {
				fmt.Print(err)
				ErrorPage(w, "Error executing template", 500)
				return
			}
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else if r.Method == http.MethodGet {
		file := "./ui/templates/reg.html"
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			ErrorPage(w, "Error parsing templates", 500)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			fmt.Print(err)
			ErrorPage(w, "Error executing template", 500)
			return
		}
	} else {
		ErrorPage(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (a *AuthHanlder) Logout(w http.ResponseWriter, r *http.Request) {
	// Delete the cookie
	cookie, err := cookies.GetCookie(r)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		ErrorPage(w, "Failed to get cookie", http.StatusInternalServerError)
		return
	}

	// Delete the session in the database
	err = a.Service.SessionService.DeleteSessionByID(cookie.Value)
	if err != nil {
		logger.GetLogger().Error("Failed to delete session in the database:")
		ErrorPage(w, "Failed to delete session in the database", http.StatusInternalServerError)
		return
	}

	// Delete the cookie
	cookies.DeleteCookie(w) // TODO check if it works

	// Redirect to /
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
