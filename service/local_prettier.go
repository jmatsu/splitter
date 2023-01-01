package service

type LocalDistributionResult struct {
	localMoveResponse
	RawJson string
}

func (r *LocalDistributionResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *LocalDistributionResult) ValueResponse() any {
	return *r
}
