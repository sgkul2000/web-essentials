package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

// Template is the template struct for echo renderer
type Template struct {
	templates *template.Template
}

// Render function for echo
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
