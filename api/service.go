package api

import (
	"context"
	"fmt"
)

type ServiceAPI service

type Service struct {
	ID      int     `json:"id"`
	Service string  `json:"service"`
	Enabled bool    `json:"enable"`
	State   string  `json:"state"`
	Pids    []int64 `json:"pids"`
}

func (s *ServiceAPI) Get(ctx context.Context, id int) (*Service, error) {
	path := fmt.Sprintf("service/id/%d", id)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	svc := &Service{}

	_, err = s.client.Do(ctx, req, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
