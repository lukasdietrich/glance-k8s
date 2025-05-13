package extension

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/lukasdietrich/glance-k8s/internal/k8s"
)

func metrics(client *k8s.Client) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		nodes, err := client.ListNodes(ctx.Request().Context())
		if err != nil {
			return err
		}

		return ctx.Render(http.StatusOK, "widgets/metrics", nodes)
	}
}

func pods(*k8s.Client) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		return echo.NewHTTPError(http.StatusNotImplemented)
	}
}
