package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
)

type firebaseAppDistributionAabInfoRequest struct {
	projectNumber string
	appId         string
}

type firebaseAppDistributionAabInfoResponse struct {
	IntegrationState appBundleIntegrationState              `json:"integrationState"`
	TestCertificate  *firebaseAppDistributionAppCertificate `json:"testCertificate"`

	RawResponse *net.HttpResponse `json:"-"`
}

func (r *firebaseAppDistributionAabInfoResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

type firebaseAppDistributionAppCertificate struct {
	Sha1   string `json:"hashSha1"`
	Sha256 string `json:"hashSha256"`
	Md5    string `json:"hashMd5"`
}

type appBundleIntegrationState = string

const (
	aabIntegrationUnspecified     appBundleIntegrationState = "AAB_INTEGRATION_STATE_UNSPECIFIED"           // Unknown
	aabIntegrationIntegrated      appBundleIntegrationState = "INTEGRATED"                                  // Available
	aabIntegrationNotLinked       appBundleIntegrationState = "PLAY_ACCOUNT_NOT_LINKED"                     // Users need to link their play store account and firebase project
	aabIntegrationNoAppFound      appBundleIntegrationState = "NO_APP_WITH_GIVEN_BUNDLE_ID_IN_PLAY_ACCOUNT" // Given apps do not register to play store. . App bundle is unavailable by spec.
	aabIntegrationNonPublished    appBundleIntegrationState = "APP_NOT_PUBLISHED"                           // Users have not published their apps yet. App bundle is unavailable by spec.
	aabIntegrationUnavailable     appBundleIntegrationState = "AAB_STATE_UNAVAILABLE"                       // Play store may have some troubles.
	aabIntegrationTermsUnaccepted appBundleIntegrationState = "PLAY_IAS_TERMS_NOT_ACCEPTED"                 // Users need to agree the terms first
)

func (p *FirebaseAppDistributionProvider) getAabInfo(request *firebaseAppDistributionAabInfoRequest) (*firebaseAppDistributionAabInfoResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/apps/%s/aabInfo", request.projectNumber, request.appId)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	resp, err := client.DoGet(p.ctx, []string{path}, nil)

	if err != nil {
		return nil, err
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&firebaseAppDistributionAabInfoResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to get aab info but something went wrong")
		} else {
			return v.(*firebaseAppDistributionAabInfoResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to get aab info")
	}
}

func (r *firebaseAppDistributionAabInfoResponse) Available() bool {
	return r.IntegrationState == aabIntegrationIntegrated
}

func checkAppBundleIntegrationState(s appBundleIntegrationState) error {
	switch s {
	case aabIntegrationIntegrated:
		firebaseAppDistributionLogger.Debug().Msgf("aab is available")
		return nil
	case aabIntegrationNonPublished:
		return errors.New(fmt.Sprintf("you have to publish apps as for app bundle uploads though, you can use apk uploads: %s", s))
	case aabIntegrationNotLinked:
		return errors.New(fmt.Sprintf("yuo have to link this firebase project with play store: %s", s))
	case aabIntegrationNoAppFound:
		return errors.New(fmt.Sprintf("this package name is not found in play store: %s", s))
	case aabIntegrationTermsUnaccepted:
		return errors.New(fmt.Sprintf("you have to accept the terms of playstore: %s", s))
	case aabIntegrationUnavailable:
		return errors.New(fmt.Sprintf("playstore currently seems down: %s", s))
	case aabIntegrationUnspecified:
		return errors.New(fmt.Sprintf(": %s", s))
	default:
		firebaseAppDistributionLogger.Warn().Msgf("unsupported aab info but we allow to continue: %s", s)
		return nil
	}
}
