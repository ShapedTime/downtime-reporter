package reader

import (
	"downtime-reporter/core"
	"downtime-reporter/transformers"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Prometheus struct {
	url         *url.URL
	queryParams *PrometheusQueryParams
}

type PrometheusQueryParams struct {
	Query    string
	Start    *time.Time
	End      *time.Time
	Step     *time.Duration
	UnixTime bool
}

type promResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string       `json:"resultType"`
		Result     []promResult `json:"result"`
	} `json:"data"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

const (
	Exceeded11kPointsError = "exceeded maximum resolution of 11,000 points per timeseries"
)

func NewPrometheusReader(promUrl string, params ...PrometheusQueryParams) (*Prometheus, error) {
	u, err := url.Parse(promUrl)
	if err != nil {
		return nil, err
	}
	apiPath, err := url.Parse("api/v1/query_range")
	if err != nil {
		return nil, err
	}
	u = u.ResolveReference(apiPath)

	reader := &Prometheus{url: u}

	if len(params) > 0 {
		err := reader.SetQueryParams(params[0])
		if err != nil {
			return nil, err
		}
	}

	return reader, nil
}

func (p *Prometheus) SetQueryParams(params PrometheusQueryParams) error {
	if params.Query == "" ||
		params.Start == nil ||
		params.End == nil ||
		params.Step == nil {
		return fmt.Errorf("all the parameters should be set")
	}

	p.queryParams = &params

	return nil
}

func (p *Prometheus) Read() (core.Result, error) {
	if p.queryParams == nil {
		return core.Result{}, fmt.Errorf("parameters should be set in order to Read")
	}

	p.setQueryAndEncode()

	resp, err := p.getPromResponse()
	if err != nil {
		return core.Result{}, fmt.Errorf("cannot get prom response: %v", err)
	}

	results, err := p.mapResponseToResults(resp)

	if err != nil {
		return core.Result{}, fmt.Errorf("cannot map prom response to results: %v", err)
	}

	return results, nil
}

func (p *Prometheus) setQueryAndEncode() {
	q := p.url.Query()
	q.Set("query", p.queryParams.Query)
	q.Set("start", fmt.Sprintf("%d", p.queryParams.Start.Unix()))
	q.Set("end", fmt.Sprintf("%d", p.queryParams.End.Unix()))
	q.Set("step", p.queryParams.Step.String())
	p.url.RawQuery = q.Encode()
}

func (p *Prometheus) getPromResponse() (*promResponse, error) {
	response, err := http.Get(p.url.String())
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
			return nil, fmt.Errorf(Exceeded11kPointsError)
		}

		return nil, fmt.Errorf("%s didn't return 200 OK but %s", p.url.String(), response.Status)
	}

	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *Prometheus) mapResponseToResults(resp *promResponse) (core.Result, error) {
	if resp.Data.ResultType != "matrix" {
		return core.Result{}, fmt.Errorf("result type isn't of type matrix: %s", resp.Data.ResultType)
	}

	if len(resp.Data.Result) == 0 {
		return core.Result{}, fmt.Errorf("no timeseries found")
	}

	// as we are only taking one metric, we should get only one results array
	res := resp.Data.Result[0]
	r := core.Result{
		Values:       make(map[string][]string),
		KeysSlice:    make([]string, 0),
		Transformers: make([]string, 0),
	}
	for _, vals := range res.Values {
		timestamp := strconv.Itoa(int(vals[0].(float64)))

		r.Values[timestamp] = []string{vals[1].(string)}
		r.KeysSlice = append(r.KeysSlice, timestamp)
	}

	// we need casting for transformers to work
	r = transformers.SortingTransformer(r).Transform()

	return r, nil
}
