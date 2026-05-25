package dynamitedb

import "errors"

// external error API to catch specific conditions.
var (
	// the specified object was not found
	ErrNotFound = errors.New("not found")
	// the object you want to insert does already exist
	ErrAlreadyExists = errors.New("already exists")
	// optimistic locking failure (operation aborted because another writer changed object mid transaction)
	ErrConcurrencyConflict = errors.New("optimistic lock failed")
)
