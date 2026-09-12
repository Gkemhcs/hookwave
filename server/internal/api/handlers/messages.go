package handlers

// Response text shared across handlers. Lives here, not in a generic utils
// package, since this is HTTP-response copy and handlers is its only consumer.
const (
	InternalServerError     = "Internal Server Error"
	InternalServerErrorDesc = "Server failed to process  the request"
	EnvironmentNotFound     = "Environment not Found"
)
