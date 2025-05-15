package storage

import "errors"

var (
	ErrURLNotFound = errors.New("task not found")
	ErrURLExists   = errors.New("task not exists")
)
