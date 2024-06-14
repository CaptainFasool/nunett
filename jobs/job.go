// Package jobs does everything job-related
package jobs

import "gitlab.com/nunet/device-management-service/models"

// Job specification
type Job struct {
	ID         string
	Name       string
	TaskGroups []TaskGroup
	// TaskGroups [][]*models.ExecutionRequest
}

// TaskGroup define tasks
type TaskGroup struct {
	Tasks []*models.ExecutionRequest
}
