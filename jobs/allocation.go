// Package jobs deals with everything job-related
package jobs

import (
	"gitlab.com/nunet/device-management-service/models"
)

// Allocation maps an execution request to a given node
type Allocation struct {
	ID     string
	JobID  string
	NodeID string

	// TODO: transform into a slice of requests
	Request *models.ExecutionRequest
}

// Resources return the allocation's total amount of resources
func (a *Allocation) Resources() *models.ExecutionResources {
	// TODO: return the sum of all execution request's resources
	return a.Request.Resources
}

// NewAllocation returns a pointer to *Allocation
func NewAllocation(id, nodeID string, req *models.ExecutionRequest) *Allocation {
	return &Allocation{
		ID:      id,
		NodeID:  nodeID,
		Request: req,
	}
}
