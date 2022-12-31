package firebase_app_distribution

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type DistributionResult struct {
	uploadResponse
	RawJson string
}

type uploadResponse struct {
}

var TableBuilder = func(w table.Writer, v any) {
	_ = v.(DistributionResult)

}
