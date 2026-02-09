package extension

import (
	"embed"
	"fmt"
	"html/template"
	"regexp"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v5"
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
		"url":                    urlTemplateFunc(),
		"icon":                   iconTemplateFunc(),
		"formatResourceQuantity": formatResourceQuantityTemplateFunc(),
	}
}

func urlTemplateFunc() func(string) template.URL {
	return func(url string) template.URL {
		return template.URL(url)
	}
}

func iconTemplateFunc() func(string) template.URL {
	pattern := regexp.MustCompile(`^((?P<shorthand>(di|si)):)(?P<name>[^\s\.]*)(\.(?P<extension>\w*))?$`)

	return func(s string) template.URL {
		if match := pattern.FindStringSubmatch(s); match != nil {
			var (
				shorthand = match[pattern.SubexpIndex("shorthand")]
				name      = match[pattern.SubexpIndex("name")]
				extension = match[pattern.SubexpIndex("extension")]
			)

			return template.URL(resolveShorthandIcon(shorthand, name, extension))
		}

		return template.URL(s)
	}
}

func resolveShorthandIcon(shorthand, name, extension string) string {
	if extension == "" {
		extension = "svg"
	}

	switch shorthand {
	case "si":
		return fmt.Sprintf("https://cdn.jsdelivr.net/npm/simple-icons@latest/icons/%[1]s.svg", name)
	case "di":
		return fmt.Sprintf("https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/%[2]s/%[1]s.%[2]s", name, extension)
	default:
		return fmt.Sprintf("unknown icon shorthand %q", shorthand)
	}
}

func formatResourceQuantityTemplateFunc() func(*resource.Quantity) template.HTML {
	return func(quantity *resource.Quantity) template.HTML {
		scaled := quantity.ScaledValue(resource.Mega)
		return template.HTML(fmt.Sprintf(`%d <span class="color-base size-h5">Mi</span>`, scaled))
	}
}
