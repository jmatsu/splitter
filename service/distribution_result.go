package service

type DistributionResult interface {
	TypedResponse() any
	RawJsonResponse() string
}
