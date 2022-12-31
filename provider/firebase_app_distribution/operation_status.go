package firebase_app_distribution

import (
	"encoding/json"
	"fmt"
	"github.com/jmatsu/splitter/internal/config"
	"github.com/pkg/errors"
	"time"
)

type getOperationStateRequest struct {
	operationName string
}

type getOperationStateResponse struct {
	OperationName string                   `json:"name"`
	Done          bool                     `json:"done"`
	Response      *v1UploadReleaseResponse `json:"response,omitempty"`
}

type v1UploadReleaseResponse struct {
	Result  string `json:"result"`
	Release release
}

// Wait until the processing in app distribution has done
func (p *Provider) waitForOperationDone(request *getOperationStateRequest) (*getOperationStateResponse, error) {
	waitTimeout := config.GetGlobalConfig().WaitTimeout()

	var retryCount int

	pipeline := make(chan *getOperationStateResponse, 1)
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

				logger.Warn().Msg("The processing of Firebase seems to be unstable. We are retrying to watch the status.")

				retryCount++
			} else if resp.Done {
				logger.Info().Msg("The processing of Firebase has done.")
				pipeline <- resp
				return
			}

			logger.Info().Msg("Waiting for the processing of Firebase...")

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

func (p *Provider) getOperationState(request *getOperationStateRequest) (*getOperationStateResponse, error) {
	path := fmt.Sprintf("/v1/%s", request.operationName)

	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	code, bytes, err := client.DoGet(p.ctx, []string{path}, nil)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get a response from operation state api")
	}

	var response getOperationStateResponse

	if 200 <= code && code < 300 {
		if err := json.Unmarshal(bytes, &response); err != nil {
			return nil, errors.Wrap(err, "cannot unmarshal operation state response")
		} else {
			return &response, nil
		}
	} else {
		return nil, errors.New(fmt.Sprintf("got %d response: %s", code, string(bytes)))
	}
}
