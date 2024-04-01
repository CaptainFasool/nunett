package docker

const (
	labelExecutorName = "nunet-executor"
	labelJobID        = "nunet-jobID"
	labelExecutionID  = "nunet-executionID"
)

func labelJobValue(executorID string, jobID string) string {
	return fmt.Sprintf("%s_%s", executorID, jobID)
}

func labelExecutionValue(executorID string, jobID string, executionID string) string {
	return fmt.Sprintf("%s_%s_%s", executorID, jobID, executionID)
}
