package workflow

// NoResult is a placeholder for when a workflow has not yet produced a final output.
var NoResult Result = nil

// Result represents the final output or data produced by a workflow execution.
type Result any
