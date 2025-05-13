package extension

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	headerContentType      = "Widget-Content-Type"
	headerContentFrameless = "Widget-Content-Frameless"
	headerTitle            = "Widget-Title"
)

func widgetContentType(contentType string) echo.MiddlewareFunc {
	return makeWidgetMiddleware(headerContentType, contentType)
}

func widgetContentFrameless(frameless bool) echo.MiddlewareFunc {
	return makeWidgetMiddleware(headerContentFrameless, strconv.FormatBool(frameless))
}

func widgetTitle(title string) echo.MiddlewareFunc {
	return makeWidgetMiddleware(headerTitle, title)
}

func makeWidgetMiddleware(name, value string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Response().Header().Add(name, value)
			return next(ctx)
		}
	}
}
