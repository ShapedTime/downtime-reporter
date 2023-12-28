package reader

import (
	"downtime-reporter/core"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type PrometheusV2 struct {
	*Prometheus
}

func NewPrometheusV2Reader(promUrl string, params ...PrometheusQueryParams) (*PrometheusV2, error) {
	u, err := url.Parse(promUrl)
	if err != nil {
		return nil, err
	}
	apiPath, err := url.Parse("api/v1/query_range")
	if err != nil {
		return nil, err
	}
	u = u.ResolveReference(apiPath)

	reader := &PrometheusV2{&Prometheus{url: u}}

	if len(params) > 0 {
		err := reader.SetQueryParams(params[0])
		if err != nil {
			return nil, err
		}
	}

	return reader, nil
}

func (p *PrometheusV2) SetQueryParams(params PrometheusQueryParams) error {
	if params.Query == "" ||
		params.Start == nil ||
		params.End == nil ||
		params.Step == nil {
		return fmt.Errorf("all the parameters should be set")
	}

	p.queryParams = &params

	return nil
}

func (p *PrometheusV2) Read() (core.Result, error) {
	if p.queryParams == nil {
		return core.Result{}, fmt.Errorf("parameters should be set in order to Read")
	}

	resp, err := getPromResponse(p.url.String(), p.queryParams.Query, p.queryParams.Step.String(), p.queryParams.Start.Unix(), p.queryParams.End.Unix())
	if err != nil {
		return core.Result{}, fmt.Errorf("cannot get prom response: %v", err)
	}

	results, err := p.mapResponseToResults(resp)

	if err != nil {
		return core.Result{}, fmt.Errorf("cannot map prom response to results: %v", err)
	}

	return results, nil
}

func setQueryAndEncode(urlStr, query, step string, startUnix, endUnix int64) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("cannot parse url: %v", err)
	}

	q := u.Query()
	q.Set("query", query)
	q.Set("start", fmt.Sprintf("%d", startUnix))
	q.Set("end", fmt.Sprintf("%d", endUnix))
	q.Set("step", step)
	u.RawQuery = q.Encode()

	return u
}

func getPromResponse(urlStr, query, step string, startUnix, endUnix int64) (*promResponse, error) {
	u := setQueryAndEncode(urlStr, query, step, startUnix, endUnix)

	response, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var resp promResponse

	if response.StatusCode != 200 {

		if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
			return nil, fmt.Errorf("cannot decode response body with status code %s: %v", response.Status, err)
		}

		if strings.Contains(resp.Error, Exceeded11kPointsError) {
			midPoint := startUnix + (endUnix-startUnix)/2
			firstHalf, err := getPromResponse(urlStr, query, step, startUnix, midPoint)
			if err != nil {
				return nil, fmt.Errorf("cannot get first half of response: %v", err)
			}

			secondHalf, err := getPromResponse(urlStr, query, step, midPoint, endUnix)
			if err != nil {
				return nil, fmt.Errorf("cannot get second half of response: %v", err)
			}

			return mergePromResponses(firstHalf, secondHalf), nil
		}

		return nil, fmt.Errorf("%s didn't return 200 OK but %s", u.String(), response.Status)
	}

	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func mergePromResponses(firstHalf *promResponse, secondHalf *promResponse) *promResponse {
	var resp promResponse

	resp.Status = firstHalf.Status
	resp.Data.ResultType = firstHalf.Data.ResultType
	resp.Data.Result = []promResult{{}}
	if len(secondHalf.Data.Result) == 0 {
		return &resp
	} else if len(firstHalf.Data.Result) == 0 {
		resp.Data.Result[0].Values = secondHalf.Data.Result[0].Values
		return &resp
	}
	resp.Data.Result[0].Values = append(firstHalf.Data.Result[0].Values, secondHalf.Data.Result[0].Values...)

	return &resp
}
