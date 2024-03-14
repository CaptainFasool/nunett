package firecracker

import (
	"github.com/samber/lo"
	"go.opentelemetry.io/otel"

	"github.com/bacalhau-project/bacalhau/pkg/telemetry"
)

var (
	firecrackerExecutorMeter = otel.GetMeterProvider().Meter("firecracker-executor")
)

var (
	ActiveExecutions = lo.Must(telemetry.NewGauge(
		firecrackerExecutorMeter,
		"firecracker_active_executions",
		"Number of active firecracker executions",
	))
)

