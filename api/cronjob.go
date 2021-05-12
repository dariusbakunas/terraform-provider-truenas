package api

import (
	"context"
	"fmt"
)

type CronjobService service

type JobSchedule struct {
	Minute string `json:"minute"`
	Hour   string `json:"hour"`
	Dom    string `json:"dom"`
	Month  string `json:"month"`
	Dow    string `json:"dow"`
}

type Job struct {
	ID          int         `json:"id"`
	User        string      `json:"user"`
	Command     string      `json:"command"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled"`
	STDOUT      bool        `json:"stdout"`
	STDERR      bool        `json:"stderr"`
	Schedule    JobSchedule `json:"schedule"`
}

func (s *CronjobService) Get(ctx context.Context, id int) (*Job, error) {
	path := fmt.Sprintf("cronjob/id/%d", id)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	job := &Job{}

	_, err = s.client.Do(ctx, req, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}
