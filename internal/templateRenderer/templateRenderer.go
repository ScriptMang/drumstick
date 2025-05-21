package templateRenderer

import (
	"html/template"
	"io"
	"log"

	"github.com/labstack/echo/v4"
)

type TemplateManager struct {
	Templates *template.Template
}

func (tm *TemplateManager) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	err := tm.Templates.ExecuteTemplate(w, name, data)

	if err != nil {
		log.Println("template not found")
	}

	return err
}
