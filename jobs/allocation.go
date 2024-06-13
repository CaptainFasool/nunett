// Package jobs deals with everything job-related
package jobs

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/nunet/device-management-service/executor"
	"gitlab.com/nunet/device-management-service/models"
)

// a jobspec will be transformed in one or multiple jobs
type job struct {
	id   string
	reqs []*models.ExecutionRequest // id of each request match job id
	executor.Executor
}

func (j *job) ExecuteAll(ctx context.Context) {
	// add logic for concurrent dispatch
	for _, req := range j.reqs {
		j.Run(ctx, req)
	}
}

// Allocation maps an execution request to a given node
type Allocation struct {
	ID      string
	NodeID  string
	Request *models.ExecutionRequest
}

// NewAllocation returns a pointer to *Allocation
func NewAllocation(nodeID string, req *models.ExecutionRequest) *Allocation {
	return &Allocation{
		ID:      "teste",
		NodeID:  nodeID,
		Request: req,
	}
}

// Allocater manages allocations
type Allocater interface {
	Allocate(ctx context.Context, req *models.ExecutionRequest) *Allocation
}

// Allocator is a form of collector for allocations
// It implements the Allocater interface
type Allocator struct {
	ch chan *models.ExecutionRequest
}

// NewAllocator returns a pointer to allocator struct
func NewAllocator() *Allocator {
	return &Allocator{
		ch: make(chan *models.ExecutionRequest),
	}
}

// Listen wait on a *Allocation channel until it receives a value
// Once a value is read, it calls Allocate method
func (a *Allocator) Listen(ctx context.Context, w io.Writer) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stopping listen...")
				break
			case <-a.ch:
				fmt.Fprintf(w, "received execution request!")
			}
		}
	}()
}

// Allocate currently implements a placeholder text
// Here we should define the mechanism to actually place the allocation (e.g. RPC calls, libp2p streams etc.)
func (a *Allocator) Allocate(_ context.Context, alloc *Allocation) {
	fmt.Printf("allocation %s: placing execution %s at node %s", alloc.ID, alloc.Request.ExecutionID, alloc.NodeID)
}
