package util

import "log"

// LogError logs the error and then exit the program
// if there is any error provided, otherwise do nothing
func LogError(err error) {
	if err != nil {
		log.Fatal("task-cli: error: ", err)
	}
}
