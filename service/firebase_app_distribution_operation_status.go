package service

import (
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/jmatsu/splitter/internal/net"
	"github.com/pkg/errors"
	"time"
)

type firebaseAppDistributionGetOperationStateRequest struct {
	operationName string
}

type FirebaseAppDistributionGetOperationStateResponse struct {
	OperationName string                                          `json:"name"`
	Done          bool                                            `json:"done"`
	Response      *FirebaseAppDistributionV1UploadReleaseResponse `json:"response"`
	RawResponse   *net.HttpResponse                               `json:"-"`
}

func (r *FirebaseAppDistributionGetOperationStateResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

var _ net.TypedHttpResponse = &FirebaseAppDistributionGetOperationStateResponse{}

type FirebaseAppDistributionV1UploadReleaseResponse struct {
	Result      string `json:"result"`
	Release     FirebaseAppDistributionReleaseFragment
	RawResponse *net.HttpResponse `json:"-"`
}

func (r *FirebaseAppDistributionV1UploadReleaseResponse) Set(v *net.HttpResponse) {
	r.RawResponse = v
}

var _ net.TypedHttpResponse = &FirebaseAppDistributionV1UploadReleaseResponse{}

// Wait until the processing in app distribution has done
func (p *FirebaseAppDistributionProvider) waitForOperationDone(request *firebaseAppDistributionGetOperationStateRequest) (*FirebaseAppDistributionGetOperationStateResponse, error) {
	waitTimeout := config.CurrentConfig().WaitTimeout()

	var retryCount int

	pipeline := make(chan *FirebaseAppDistributionGetOperationStateResponse, 1)
	stopper := make(chan error, 1)

	defer func() {
		close(pipeline)
		close(stopper)
	}()

	go func() {
		for {
			if resp, err := p.getOperationState(request); err != nil {
				// experimental
				if retryCount >= 5 {
					stopper <- errors.Wrap(err, "retry limit exceeded while waiting for the operation")
					return
				}

				firebaseAppDistributionLogger.Warn().Msg("The processing of Firebase seems to be unstable. We are retrying to watch the status.")

				retryCount++
			} else if resp.Done {
				firebaseAppDistributionLogger.Info().Msg("The processing of Firebase has done.")
				pipeline <- resp
				return
			}

			firebaseAppDistributionLogger.Info().Msg("Waiting for the processing of Firebase...")

			time.Sleep(5 * time.Second) // experimental
		}
	}()

	select {
	case err := <-stopper:
		return nil, err
	case resp := <-pipeline:
		return resp, nil
	case <-time.After(waitTimeout):
		return nil, errors.New("time limit exceeded while waiting for the operation")
	}
}

func (p *FirebaseAppDistributionProvider) getOperationState(request *firebaseAppDistributionGetOperationStateRequest) (*FirebaseAppDistributionGetOperationStateResponse, error) {
	path := fmt.Sprintf("/v1/%s", request.operationName)

	client := p.client.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	resp, err := client.DoGet(p.ctx, []string{path}, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get a response from operation state api")
	}

	if resp.Successful() {
		if v, err := resp.ParseJson(&FirebaseAppDistributionGetOperationStateResponse{}); err != nil {
			return nil, errors.Wrap(err, "succeeded to monitor the operation state but something went wrong")
		} else {
			return v.(*FirebaseAppDistributionGetOperationStateResponse), nil
		}
	} else {
		return nil, errors.Wrap(resp.Err(), "failed to monitor the operation state")
	}
}
