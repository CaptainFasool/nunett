// Package service contains abstraction for job runners.
package service

// JobRunner is kind of base class for all the runner.
// Note: When adding new methods for other comcrete class (interface implementations),
// make sure methods are common to all the runners.
type JobRunner interface {
	Run() error
	Stop() error
	IsRunning() bool
	IsStopped() bool
}
