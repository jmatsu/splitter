package firebase_app_distribution

import (
	"encoding/json"
	"fmt"
)

type aabInfoRequest struct {
	projectNumber string
	appId         string
}

type aabInfoResponse struct {
	IntegrationState integrationState `json:"integrationState"`
	TestCertificate  *certificate     `json:"testCertificate,omitempty"`
}

type certificate struct {
	Sha1   string `json:"hashSha1"`
	Sha256 string `json:"hashSha256"`
	Md5    string `json:"hashMd5"`
}

type integrationState = string

const (
	aabIntegrationUnspecified     integrationState = "AAB_INTEGRATION_STATE_UNSPECIFIED"
	aabIntegrationIntegrated      integrationState = "INTEGRATED"
	aabIntegrationNotLinked       integrationState = "PLAY_ACCOUNT_NOT_LINKED"
	aabIntegrationNoAppFound      integrationState = "NO_APP_WITH_GIVEN_BUNDLE_ID_IN_PLAY_ACCOUNT"
	aabIntegrationNonPublished    integrationState = "APP_NOT_PUBLISHED"
	aabIntegrationUnavailable     integrationState = "AAB_STATE_UNAVAILABLE"
	aabIntegrationTermsUnaccepted integrationState = "PLAY_IAS_TERMS_NOT_ACCEPTED"
)

func (p *Provider) getAabInfo(request *aabInfoRequest) (*aabInfoResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/apps/%s/aabInfo", request.projectNumber, request.appId)

	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	code, bytes, err := client.DoGet(p.ctx, []string{path}, nil)

	if err != nil {
		return nil, err
	}

	var response aabInfoResponse

	if 200 <= code && code < 300 {
		if err := json.Unmarshal(bytes, &response); err != nil {
			return nil, fmt.Errorf(": %v", err)
		} else {
			return &response, nil
		}
	} else {
		return nil, fmt.Errorf("got %d response: %s", code, string(bytes))
	}
}

func checkIntegrationState(s integrationState) error {
	switch s {
	case aabIntegrationIntegrated:
		logger.Debug().Msgf("aab is available")
		return nil
	case aabIntegrationNonPublished:
		return fmt.Errorf("you have to publish apps as for app bundle uploads though, you can use apk uploads: %s", s)
	case aabIntegrationNotLinked:
		return fmt.Errorf("yuo have to link this firebase project with play store: %s", s)
	case aabIntegrationNoAppFound:
		return fmt.Errorf("this package name is not found in play store: %s", s)
	case aabIntegrationTermsUnaccepted:
		return fmt.Errorf("you have to accept the terms of playstore: %s", s)
	case aabIntegrationUnavailable:
		return fmt.Errorf("playstore currently seems down: %s", s)
	case aabIntegrationUnspecified:
		return fmt.Errorf(": %s", s)
	default:
		logger.Warn().Msgf("unsupported aab info but we allow to continue: %s", s)
		return nil
	}
}
