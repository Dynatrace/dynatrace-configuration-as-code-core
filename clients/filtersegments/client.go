package filtersegments

import (
	"context"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
)

//go:generate mockgen -source client.go -package=filtersegments -destination=client_mock.go
type client interface {
	Create(ctx context.Context, data []byte) (*http.Response, error)
	Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
	List(ctx context.Context) (*http.Response, error)
	Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string) (*http.Response, error)
}
