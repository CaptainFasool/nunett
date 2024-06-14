package internal

// Shutdown logs the shutdown reason and signals the ShutdownChan
func Shutdown(message string) {
	zlog.Sugar().Infof("Shutdown initiated: %s", message)
	close(ShutdownChan)
}
