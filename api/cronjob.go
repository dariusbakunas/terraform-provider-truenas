package api

import (
	"context"
	"fmt"
	"net/url"
)

type CronjobService service

type JobSchedule struct {
	Minute string `json:"minute,omitempty"`
	Hour   string `json:"hour,omitempty"`
	Dom    string `json:"dom,omitempty"`
	Month  string `json:"month,omitempty"`
	Dow    string `json:"dow,omitempty"`
}

type Job struct {
	ID          int         `json:"id"`
	User        string      `json:"user"`
	Command     string      `json:"command"`
	Description string      `json:"description,omitempty"`
	Enabled     bool        `json:"enabled"`
	STDOUT      bool        `json:"stdout"`
	STDERR      bool        `json:"stderr"`
	Schedule    JobSchedule `json:"schedule"`
}

type JobInput struct {
	User        string       `json:"user"`
	Command     string       `json:"command"`
	Description string       `json:"description,omitempty"`
	Enabled     *bool        `json:"enabled,omitempty"`
	STDOUT      *bool        `json:"stdout,omitempty"`
	STDERR      *bool        `json:"stderr,omitempty"`
	Schedule    *JobSchedule `json:"schedule,omitempty"`
}

func (s *CronjobService) Get(ctx context.Context, id string) (*Job, error) {
	path := fmt.Sprintf("cronjob/id/%s", id)
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

func (s *CronjobService) Create(ctx context.Context, dataset *JobInput) (*Job, error) {
	path := "cronjob"

	req, err := s.client.NewRequest("POST", path, dataset)

	if err != nil {
		return nil, err
	}

	j := &Job{}

	_, err = s.client.Do(ctx, req, j)

	if err != nil {
		return nil, err
	}

	return j, nil
}

func (s *CronjobService) Update(ctx context.Context, id string, dataset *JobInput) (*Job, error) {
	path := fmt.Sprintf("cronjob/id/%s", id)
	req, err := s.client.NewRequest("PUT", path, dataset)

	if err != nil {
		return nil, err
	}

	j := &Job{}

	_, err = s.client.Do(ctx, req, j)

	if err != nil {
		return nil, err
	}

	return j, nil
}

func (s *CronjobService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("cronjob/id/%s", url.QueryEscape(id))

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
