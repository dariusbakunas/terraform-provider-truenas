package api

import (
	"context"
)

// PoolService handles communication with the pool related
// methods of the TrueNAS API.
type PoolService service

type Pool struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func (s *PoolService) List(ctx context.Context, opts *ListOptions) ([]Pool, error) {
	path := "pool"

	path, err := addOptions(path, opts)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var pools []Pool
	_, err = s.client.Do(ctx, req, &pools)
	if err != nil {
		return nil, err
	}

	return pools, nil
}
