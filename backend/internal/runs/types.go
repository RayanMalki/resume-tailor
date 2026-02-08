package runs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type Run struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ResumeID        uuid.UUID
	JobText         string
	ProjectControls []ProjectControl
	Status          Status
	ErrorMessage    *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProjectControlMode string

const (
	ProjectControlPinned  ProjectControlMode = "pinned"
	ProjectControlAuto    ProjectControlMode = "auto"
	ProjectControlExclude ProjectControlMode = "exclude"
)

type ProjectControl struct {
	Name string             `json:"name"`
	Mode ProjectControlMode `json:"mode"`
}

var (
	ErrRunNotFound = errors.New("run failed")
	ErrForbidden   = errors.New("forbidden")
	ErrBadInput    = errors.New("bad input")
)
