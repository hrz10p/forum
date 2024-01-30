package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	// Создаем кэш шаблонов.
	tmplCache, err := NewTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	// Извлекаем шаблон из кэша по его имени.
	ts, ok := tmplCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		ServerError(w, err)
		return
	}

	// Создаем буфер для записи рендеринга.
	buf := new(bytes.Buffer)

	// Выполняем рендеринг шаблона и записываем результат в буфер.
	err = ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		ServerError(w, err)
		return
	}

	// Устанавливаем статус ответа и отправляем рендеринг клиенту.
	w.WriteHeader(status)
	buf.WriteTo(w)
}

type HandlerError struct {
	ErrorCode int
	ErrorMsg  string
}

// ServerError обрабатывает внутренние ошибки сервера, записывая их в логи и отображая пользователю страницу ошибки.
func ServerError(w http.ResponseWriter, err error) {
	//trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	errorForm := HandlerError{
		ErrorCode: http.StatusInternalServerError,
		ErrorMsg:  http.StatusText(http.StatusInternalServerError),
	}

	tmpl, err := template.ParseFiles("./ui/html/error/error.html")
	if err != nil {

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(errorForm.ErrorCode)

	if err := tmpl.Execute(w, errorForm); err != nil {

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// ClientError обрабатывает ошибки, связанные с запросами клиентов, отображая пользователю страницу ошибки.
func ClientError(w http.ResponseWriter, status int) {
	errorForm := HandlerError{
		ErrorCode: status,
		ErrorMsg:  http.StatusText(status),
	}

	tmpl, err := template.ParseFiles("./ui/html/error/error.html")
	if err != nil {

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if err := tmpl.Execute(w, errorForm); err != nil {

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// NotFound обрабатывает ошибку "страница не найдена" и отображает пользователю соответствующую страницу ошибки.
func NotFound(w http.ResponseWriter) {
	ClientError(w, http.StatusNotFound)
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
