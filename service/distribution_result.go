package service

type DistributionResult interface {
	ValueResponse() any
	RawJsonResponse() string
}
