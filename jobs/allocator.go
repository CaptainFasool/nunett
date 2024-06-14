// Package jobs do everything job-related
package jobs

import (
	"context"
	"fmt"
	"io"

	"gitlab.com/nunet/device-management-service/models"
)

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
				// TODO: actually make the allocation
			}
		}
	}()
}

// Allocate currently implements a placeholder text
// Here we should define the mechanism to actually place the allocation (e.g. RPC calls, libp2p streams etc.)
func (a *Allocator) Allocate(_ context.Context, req *models.ExecutionRequest) *Allocation {
	return &Allocation{
		Request: req,
	}
}
