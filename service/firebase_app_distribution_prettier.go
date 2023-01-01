package service

type FirebaseAppDistributionDistributionResult struct {
	Result  string
	Release firebaseAppDistributionRelease
	AabInfo *firebaseAppDistributionAabInfoResponse
	RawJson string
}

type firebaseAppDistributionUploadResponse struct {
	OperationName string `json:"name"`
}

func (r *FirebaseAppDistributionDistributionResult) RawJsonResponse() string {
	return r.RawJson
}

func (r *FirebaseAppDistributionDistributionResult) TypedResponse() any {
	return *r
}
