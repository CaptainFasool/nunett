package models

type IntegrationLevel string
type LifeCycle string
type PluginExecutor string

const (
	Standalone IntegrationLevel = "standalone"
	Integrated IntegrationLevel = "integrated"

	SideCar  LifeCycle = "sidecar"
	OnDemand LifeCycle = "ondemand"

	Docker PluginExecutor = "docker"
)

type Plugin struct {
	Name                  string                      `json:"name"`
	LifeCycle             LifeCycle                   `json:"life_cycle"`
	IntegrationLevel      IntegrationLevel            `json:"integration_level"`
	Executor              PluginExecutor              `json:"executor"`
	AverageResourcesUsage ServiceResourceRequirements `json:"average_resources_usage"`
}
