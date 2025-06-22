# Akamai API Helper (Go)

A lightweight Go module for making authenticated calls to Akamai {OPEN} APIs using the Akamai EdgeGrid v11 authentication SDK

## Features

* **EdgeGrid v11** authentication via `github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/edgegrid`
* Automatic **retry logic** (default **10** attempts, **5s** between tries)
* Captures every attempt’s **HTTP status code** and **error/response body**
* Accepts **arbitrary request bodies** (any Go struct or `map[string]interface{}`)
* Parses responses into **caller-provided structs** for flexible JSON unmarshalling

## Installation

```bash
# Initialize or update your module
go get github.com/iangetz/akamaiapi
```

## Usage

```go
import (
	"fmt"

	"github.com/iangetz/akamaiapi"
)

// Define your response structs matching the response format of the applicable Akamai {OPEN} API
// Example struct for 'Traffic by Hostname' API (https://techdocs.akamai.com/reporting/reference/delivery-traffic-current)
// Desired response only needs the data[] element

type HostnameTraffic struct {
	Hostname                 string  `json:"hostname"`
	EdgeHitsSum              int64   `json:"edgeHitsSum"`
	MidgressHitsSum          int64   `json:"midgressHitsSum"`
	OriginHitsSum            int64   `json:"originHitsSum"`
	OffloadedHitsPercentage  float64 `json:"offloadedHitsPercentage"`
	EdgeBytesSum             int64   `json:"edgeBytesSum"`
	MidgressBytesSum         int64   `json:"midgressBytesSum"`
	OriginBytesSum           int64   `json:"originBytesSum"`
	OffloadedBytesPercentage float64 `json:"offloadedBytesPercentage"`
}

type TrafficResponse struct {
	Data []HostnameTraffic `json:"data"`
}

func main() {
	// Build request body or use map[string]interface{}
	body := map[string]interface{}{
		"dimensions": []string{"hostname"},
		"metrics": []string{
			"edgeHitsSum", "midgressHitsSum", "originHitsSum",
			"offloadedHitsPercentage", "edgeBytesSum", "midgressBytesSum",
			"originBytesSum", "offloadedBytesPercentage",
		},
		"sortBys": []map[string]string{
			{"name": "hostname", "sortOrder": "ASCENDING"},
		},
		"limit": 5,
	}

	var respModel TrafficResponse

	apiResp, err := akamaiapi.DoRequest(&akamaiapi.RequestConfig{
		EdgercPath:    "/path/to/.edgerc",
		Section:       "section name",
		Path:          "/reporting-api/v2/reports/delivery/traffic/current/data",
		Method:        "POST",
		Params:        "accountSwitchKey=<abc123>&start=<YYYY-MM-DDTHH%3AMM%3ASSZ>&<end=YYYY-MM-DDTHH%3AMM%3ASSZ>",
		Body:          body,
		Headers:       nil, // optional
		ResponseModel: &respModel,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Status codes:", apiResp.StatusCodes)
	fmt.Println("Errors:", apiResp.Errors)

	if len(respModel.Data) > 0 {
		fmt.Printf("First hostname: %s with %d hits\n",
			respModel.Data[0].Hostname,
			respModel.Data[0].EdgeHitsSum)
	}
}
```

## License

Apache License 2.0 © Ian Getz