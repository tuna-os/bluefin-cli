package commands

// Result captures the outcome of a controller command execution.
type Result struct {
	Success bool
	Message string
	Data    any
}
