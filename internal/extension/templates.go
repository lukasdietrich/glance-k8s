package extension

import (
	"embed"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:embed templates
var templatesFs embed.FS

func mustTemplates() echo.Renderer {
	return &echo.TemplateRenderer{
		Template: template.Must(
			template.New("").
				Funcs(sprig.FuncMap()).
				Funcs(templateFuncs()).
				ParseFS(templatesFs,
					"*/*.html",
					"*/*/*.html",
				),
		),
	}
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatResourceQuantity": func(quantity *resource.Quantity) template.HTML {
			scaled := quantity.ScaledValue(resource.Mega)
			return template.HTML(fmt.Sprintf(`%d <span class="color-base size-h5">Mi</span>`, scaled))
		},
	}
}
