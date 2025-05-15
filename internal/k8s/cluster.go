package k8s

import "github.com/lukasdietrich/glance-k8s/internal/k8s/api"

type Cluster struct {
	client *api.Client
}

func Connect() (*Cluster, error) {
	client, err := api.Connect()
	if err != nil {
		return nil, err
	}

	return &Cluster{client: client}, nil
}
