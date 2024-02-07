package main

import (
	"html/template"
	"net/http"
)

type Error struct {
	Message string
	Code    int
}

func ErrorPage(w http.ResponseWriter, str string, code int) {
	file := "./ui/templates/error.html"
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		http.Error(w, "Error parsing templates \n "+str, 500)
		return
	}

	tmpl.Execute(w, Error{Message: str, Code: code})
}
