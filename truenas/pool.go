package truenas

import (
	"context"
	"net/http"
)

// PoolService handles communication with the pool related
// methods of the TrueNAS API.
type PoolService service

type Pool struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Path *string `json:"path,omitempty"`
}

func (s *PoolService) List(ctx context.Context, opts *ListOptions) ([]*Pool, *http.Response, error) {
	path := "pool"

	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var pools []*Pool
	resp, err := s.client.Do(ctx, req, &pools)
	if err != nil {
		return nil, resp, err
	}

	return pools, resp, nil
}
