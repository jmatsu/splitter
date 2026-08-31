package service

type DeployResult interface {
	ValueResponse() any
	RawJsonResponse() string

	// NormalizedResponse is a service-agnostic view of this result for the dumped artifact.
	NormalizedResponse() NormalizedResult
}
