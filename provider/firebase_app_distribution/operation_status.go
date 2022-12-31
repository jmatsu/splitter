package firebase_app_distribution

import (
	"encoding/json"
	"fmt"
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
	Release struct {
		Name           string `json:"name"`
		DisplayVersion string `json:"displayVersion"`
		BuildVersion   string `json:"buildVersion"`
		CreatedAt      string `json:"createTime"`
	}
}

func (p *Provider) waitForOperationDone(request *getOperationStateRequest) (*getOperationStateResponse, error) {
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
					stopper <- fmt.Errorf("retry limit exceeded while waiting for the operation: %v", err)
					return
				}

				retryCount++
			} else if resp.Done {
				pipeline <- resp
				return
			}

			time.Sleep(5 * time.Second) // experimental
		}
	}()

	select {
	case err := <-stopper:
		return nil, err
	case resp := <-pipeline:
		return resp, nil
	case <-time.After(5 * time.Minute): // TODO it's better to be flexible
		return nil, fmt.Errorf("time limit exceeded while waiting for the opration")
	}
}

func (p *Provider) getOperationState(request *getOperationStateRequest) (*getOperationStateResponse, error) {
	path := fmt.Sprintf("/v1/%s", request.operationName)

	client := baseClient.WithHeaders(map[string][]string{
		"Authorization": {fmt.Sprintf("Bearer %s", p.AccessToken)},
	})

	code, bytes, err := client.DoGet(p.ctx, []string{path}, nil)

	if err != nil {
		return nil, err
	}

	var response getOperationStateResponse

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
