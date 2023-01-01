package service

type DeployResult interface {
	ValueResponse() any
	RawJsonResponse() string
}
