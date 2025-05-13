package extension

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/lukasdietrich/glance-k8s/internal/k8s"
)

func New(client *k8s.Client) http.Handler {
	r := echo.New()
	r.Renderer = mustTemplates()

	r.Use(logError())
	r.Use(middleware.Recover())
	r.Use(middleware.Gzip())

	r.GET("/healthz", health())

	e := r.Group("/extension")

	e.Use(widgetContentType("html"))
	e.Use(widgetContentFrameless(false))

	e.GET("/metrics", metrics(client), widgetTitle("Kubernetes Metrics"))
	e.GET("/pods", pods(client), widgetTitle("Kubernetes Pods"))

	return r
}

func logError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			err := next(ctx)
			if err != nil {
				slog.Error("error during request",
					slog.String("url", ctx.Request().RequestURI),
					slog.Any("err", err),
				)
			}

			return err
		}
	}
}

func health() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	}
}
