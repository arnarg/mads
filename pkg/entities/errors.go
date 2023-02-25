package entities

import "fmt"

type PodmanAPIError struct {
	Cause    string `json:"cause"`
	Message  string `json:"message"`
	Response int64  `json:"response"`
}

func (e *PodmanAPIError) Error() string {
	return fmt.Sprintf("%s: %s (%d)", e.Cause, e.Message, e.Response)
}
