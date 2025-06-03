package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/lukasdietrich/glance-k8s/internal/k8s/api"
)

const (
	aName        = "glance/name"
	aIcon        = "glance/icon"
	aUrl         = "glance/url"
	aSameTab     = "glance/same-tab"
	aDescription = "glance/description"
	aId          = "glance/id"
	aParent      = "glance/parent"
)

type AppSlice []*App

func (a AppSlice) Len() int {
	return len(a)
}

func (a AppSlice) Less(i, j int) bool {
	iName, jName := a[i].Name(), a[j].Name()
	return strings.ToLower(iName) < strings.ToLower(jName)
}

func (a AppSlice) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}

type App struct {
	Annotations  map[string]string
	Ingress      *api.Ingress
	Workload     Workload
	Dependencies WorkloadSlice
}

func (a *App) Name() string {
	if title, ok := a.Annotations[aName]; ok {
		return title
	}

	return cases.Title(language.English).String(a.Workload.GetName())
}

func (a *App) Icon() string {
	if icon, ok := a.Annotations[aIcon]; ok {
		return icon
	}

	return "di:kubernetes"
}

func (a *App) Url() string {
	if url, ok := a.Annotations[aUrl]; ok {
		return url
	}

	if a.Ingress != nil {
		url := buildIngressUrl(a.Ingress)
		return url.String()
	}

	return ""
}

func buildIngressUrl(ingress *api.Ingress) url.URL {
	host, path := ingressHostPath(ingress)

	return url.URL{
		Scheme: ingressScheme(ingress),
		Host:   host,
		Path:   path,
	}
}

func ingressScheme(ingress *api.Ingress) string {
	if len(ingress.Spec.TLS) > 0 {
		return "https"
	}

	return "http"
}

func ingressHostPath(ingress *api.Ingress) (host, path string) {
	if rules := ingress.Spec.Rules; len(rules) > 0 {
		rule := rules[0]

		if http := rule.HTTP; http != nil {
			if paths := http.Paths; len(paths) > 0 {
				shortestPath := lo.MinBy(paths, func(a, b api.HTTPIngressPath) bool {
					return len(a.Path) < len(b.Path)
				})

				return rule.Host, shortestPath.Path
			}
		}
	}

	return "", ""
}

func (a *App) SameTab() bool {
	return a.Annotations[aSameTab] == "true"
}

func (a *App) Description() string {
	return a.Annotations[aDescription]
}

func (a *App) Ready() bool {
	if !a.Workload.GetStatus().Ready() {
		return false
	}

	for _, dependency := range a.Dependencies {
		if !dependency.GetStatus().Ready() {
			return false
		}
	}

	return true
}

type AppsOptions struct {
	HidePattern []string
	ShowIf      []string
}

func (c *Cluster) Apps(ctx context.Context, opts AppsOptions) (AppSlice, error) {
	workloads, err := c.workloads(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch workloads: %w", err)
	}

	if workloads, err = filterWorkloads(workloads, opts); err != nil {
		return nil, fmt.Errorf("could not filter workloads: %w", err)
	}

	services, err := c.client.Services(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch services: %w", err)
	}

	ingresses, err := c.client.Ingresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ingresses: %w", err)
	}

	findIngress := makeIngressFinder(workloads, services, ingresses)

	apps := groupApps(workloads)
	for _, app := range apps {
		sort.Stable(app.Dependencies)

		if app.Workload == nil && len(app.Dependencies) > 0 {
			app.Workload = app.Dependencies[0]
			app.Dependencies = app.Dependencies[1:]

			slog.Debug("app does not have a defined parent workload. promoting first dependency",
				slog.Group("workload",
					slog.String("namespace", app.Workload.GetNamespace()),
					slog.String("name", app.Workload.GetName()),
				),
			)
		}

		app.Annotations = app.Workload.GetAnnotations()

		if ingress, ok := findIngress(append(WorkloadSlice{app.Workload}, app.Dependencies...)...); ok {
			app.Ingress = ingress
			app.Annotations = lo.Assign(app.Annotations, ingress.GetAnnotations())
		}
	}

	sort.Stable(apps)
	return apps, nil
}

func filterWorkloads(workloads WorkloadSlice, opts AppsOptions) (WorkloadSlice, error) {
	if len(opts.HidePattern) == 0 && len(opts.ShowIf) == 0 {
		return workloads, nil
	}

	hidePatternFilter, err := buildHidePatternFilterFunc(opts.HidePattern)
	if err != nil {
		return nil, fmt.Errorf("could not build hide-pattern filter: %w", err)
	}

	showIfFilter, err := buildShowIfFilterFunc(opts.ShowIf)
	if err != nil {
		return nil, fmt.Errorf("could not build show-if filter: %w", err)
	}

	filter := func(workload Workload, _ int) bool {
		return hidePatternFilter(workload) && showIfFilter(workload)
	}

	return lo.Filter(workloads, filter), nil
}

func buildHidePatternFilterFunc(patterns []string) (func(Workload) bool, error) {
	if len(patterns) > 0 {
		slog.Warn("hide-pattern is deprecated, use show-if instead")
	}

	hidePatterns := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err != nil {
			return nil, err
		} else {
			hidePatterns[i] = re
		}
	}

	filterFunc := func(workload Workload) bool {
		fullname := resourceFullname(workload)
		for _, hidePattern := range hidePatterns {
			if hidePattern.MatchString(fullname) {
				return false
			}
		}

		return true
	}

	return filterFunc, nil
}

func buildShowIfFilterFunc(expressions []string) (func(Workload) bool, error) {
	programs := make([]*vm.Program, len(expressions))
	for i, expression := range expressions {
		program, err := expr.Compile(expression, expr.AsBool(), expr.WarnOnAny())
		if err != nil {
			return nil, err
		}

		programs[i] = program
	}

	filterFunc := func(workload Workload) bool {
		env := struct {
			Name        string            `expr:"name"`
			Namespace   string            `expr:"namespace"`
			Annotations map[string]string `expr:"annotations"`
		}{
			Name:        workload.GetName(),
			Namespace:   workload.GetNamespace(),
			Annotations: workload.GetAnnotations(),
		}

		for _, program := range programs {
			output, err := expr.Run(program, &env)
			if err != nil {
				slog.Error("could not evaluate expression", slog.Any("err", err))
				return false
			}

			if !output.(bool) {
				return false
			}
		}

		return true
	}

	return filterFunc, nil
}

func groupApps(workloads WorkloadSlice) AppSlice {
	var apps AppSlice
	mappedApps := make(map[string]*App)

	slog.Debug("grouping workloads",
		slog.Int("workloads", len(workloads)),
	)

	for _, workload := range workloads {
		annoations := workload.GetAnnotations()

		if id, ok := annoations[aId]; ok {
			slog.Debug("workload is parent of group",
				slog.String("namespace", workload.GetNamespace()),
				slog.String("name", workload.GetName()),
				slog.String("id", id),
			)

			if app, ok := mappedApps[id]; ok {
				app.Workload = workload
			} else {
				mappedApps[id] = &App{Workload: workload}
			}
		} else if parent, ok := annoations[aParent]; ok {
			slog.Debug("workload is dependency of group",
				slog.String("namespace", workload.GetNamespace()),
				slog.String("name", workload.GetName()),
				slog.String("parent", parent),
			)

			if app, ok := mappedApps[parent]; ok {
				app.Dependencies = append(app.Dependencies, workload)
			} else {
				mappedApps[parent] = &App{Dependencies: WorkloadSlice{workload}}
			}
		} else {
			slog.Debug("workload is not part of a group",
				slog.String("namespace", workload.GetNamespace()),
				slog.String("name", workload.GetName()),
			)

			apps = append(apps, &App{Workload: workload})
		}
	}

	apps = append(apps, lo.Values(mappedApps)...)
	slog.Debug("grouped apps",
		slog.Int("workloads", len(workloads)),
		slog.Int("apps", len(apps)),
	)

	return apps
}

type ingressFinderFunc func(workload ...Workload) (*api.Ingress, bool)

func makeIngressFinder(workloads WorkloadSlice, services []api.Service, ingresses []api.Ingress) ingressFinderFunc {
	serviceToIngressMap := make(map[string]api.Ingress)
	workloadToIngressMap := make(map[string]api.Ingress)

	for _, service := range services {
		for _, ingress := range ingresses {
			if isIngressForService(ingress, service) {
				slog.Debug("found ingress for service",
					slog.Group("ingress",
						slog.String("namespace", ingress.GetNamespace()),
						slog.String("name", ingress.GetName()),
					),
					slog.Group("service",
						slog.String("namespace", service.GetNamespace()),
						slog.String("name", service.GetName()),
					),
				)

				serviceToIngressMap[resourceFullname(&service)] = ingress
			}
		}
	}

	for _, workload := range workloads {
		for _, service := range services {
			if isServiceForWorkload(service, workload) {
				slog.Debug("found service for workload",
					slog.Group("service",
						slog.String("namespace", service.GetNamespace()),
						slog.String("name", service.GetName()),
					),
					slog.Group("workload",
						slog.String("namespace", workload.GetNamespace()),
						slog.String("name", workload.GetName()),
					),
				)

				if ingress, ok := serviceToIngressMap[resourceFullname(&service)]; ok {
					slog.Debug("found ingress for workload",
						slog.Group("ingress",
							slog.String("namespace", ingress.GetNamespace()),
							slog.String("name", ingress.GetName()),
						),
						slog.Group("workload",
							slog.String("namespace", workload.GetNamespace()),
							slog.String("name", workload.GetName()),
						),
					)

					workloadToIngressMap[resourceFullname(workload)] = ingress
				}
			}
		}
	}

	return func(workloads ...Workload) (*api.Ingress, bool) {
		for _, workload := range workloads {
			if ingress, ok := workloadToIngressMap[resourceFullname(workload)]; ok {
				return &ingress, true
			}
		}

		return nil, false
	}
}

func resourceFullname(resource interface {
	GetNamespace() string
	GetName() string
}) string {
	return fmt.Sprintf("%s/%s", resource.GetNamespace(), resource.GetName())
}

func isServiceForWorkload(service api.Service, workload Workload) bool {
	if service.GetNamespace() != workload.GetNamespace() {
		return false
	}

	labels := workload.GetSpec().Template.Labels
	selector := service.Spec.Selector

	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}

	return true
}

func isIngressForService(ingress api.Ingress, service api.Service) bool {
	if ingress.GetNamespace() != service.GetNamespace() {
		return false
	}

	if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service.Name == service.Name {
		return true
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service.Name == service.Name {
					return true
				}
			}
		}
	}

	return false
}
