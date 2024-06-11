package internal

import (
	"log"
)

// Shutdown logs the shutdown reason and signals the ShutdownChan
func Shutdown(message string) {
	log.Println("Shutdown initiated:", message)
	close(ShutdownChan)
}
