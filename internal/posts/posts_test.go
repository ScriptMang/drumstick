package posts

import (
	"fmt"
	"scriptmang/drumstick/internal/templateRenderer"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func dict(vals ...any) (map[string]any, error) {
	if len(vals)%2 != 0 {
		return nil, fmt.Errorf("dictionary requires an even number of arguments")
	}

	m := make(map[string]any, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		ky, ok := vals[i].(string)
		if !ok {
			return nil, fmt.Errorf("dictionary keys must be string type")
		}
		m[ky] = vals[i+1]
	}
	return m, nil
}

func setupEchoClient() *echo.Echo {
	tm := &templateRenderer.TemplateManager{
		Templates: template.Must(template.New("").Funcs(template.FuncMap{
			"dict": dict,
		}).ParseGlob("../../ui/html/pages/*[^#?!|].tmpl")),
	}

	r := echo.New()

	r.Use(middleware.SecureWithConfig(middleware.DefaultSecureConfig))
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.Renderer = tm

	return r
}
