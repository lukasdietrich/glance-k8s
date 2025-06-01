package extension

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/lukasdietrich/glance-k8s/internal/k8s"
)

func nodes(cluster *k8s.Cluster) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		nodes, err := cluster.Nodes(ctx.Request().Context())
		if err != nil {
			return err
		}

		return ctx.Render(http.StatusOK, "widgets/nodes", nodes)
	}
}

func apps(cluster *k8s.Cluster) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var req struct {
			HidePattern []string `query:"hide-pattern"`
			ShowIf      []string `query:"show-if"`
		}

		if err := ctx.Bind(&req); err != nil {
			return err
		}

		apps, err := cluster.Apps(ctx.Request().Context(), k8s.AppsOptions(req))
		if err != nil {
			return err
		}

		return ctx.Render(http.StatusOK, "widgets/apps", apps)
	}
}
