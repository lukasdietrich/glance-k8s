package extension

import (
	"embed"
	"fmt"
	"html/template"
	"strings"

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

		"icon": func(s string) template.URL {
			prefix, icon, _ := strings.Cut(s, ":")

			switch prefix {
			case "si":
				return template.URL("https://cdn.jsdelivr.net/npm/simple-icons@latest/icons/" + icon + ".svg")
			case "di":
				return template.URL("https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/svg/" + icon + ".svg")
			default:
				return template.URL(s)
			}
		},

		"url": func(url string) template.URL {
			return template.URL(url)
		},
	}
}
