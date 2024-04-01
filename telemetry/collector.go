package telemetry

type Collector interface {
	observeEvent()
	getObservedLevel()
	getEndpoint() 
}

type FileCollector struct {
	logFile string	 
}

type DatabaseCollector struct {
	databaseEndpoint string
}

type OpenTelemetryCollector struct {
	otEndpoint string
}

type ReputationCollector struct {

}
