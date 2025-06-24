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
// In this case, desired response only needs select elements from the data[] array

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
	// Build request body
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
	// For APIs that require an empty body ({}), simply use:
	// body := map[string]interface{}{}

	var respModel TrafficResponse

	// Make the API request using this Akamai API helper
	// Report start/end dates should be encoded ISO-8601 format (E.g. "2025-04-15T00%3A00%3A00Z")
	apiResp, err := akamaiapi.DoRequest(&akamaiapi.RequestConfig{
		EdgercPath:    "/path/to/your/.edgerc",
		Section:       ".edgerc section name",
		Path:          "/reporting-api/v2/reports/delivery/traffic/current/data",
		Method:        "POST",
		Params:        "accountSwitchKey=<key>&start=<timestamp>&end=<timestamp>",
		Body:          body,
		Headers:       nil, // optional
		ResponseModel: &respModel,
	})
	if err != nil {
		panic(err)
	}

	// Validate API response
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