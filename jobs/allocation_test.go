package jobs

import (
	"bytes"
	"context"
	"testing"

	"gitlab.com/nunet/device-management-service/models"
)

func TestAllocation(t *testing.T) {
	ctx := context.Background()
	buf := new(bytes.Buffer)

	allocator := NewAllocator()
	allocator.Listen(ctx, buf)

	exec := models.ExecutionRequest{}
	allocator.ch <- &exec

	got := buf.String()
	want := "received execution request!"

	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
