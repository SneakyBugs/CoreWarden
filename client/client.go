package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"text/template"
	"time"

	"git.houseofkummer.com/lior/home-dns/api/services/records"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"github.com/miekg/dns"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientOptions struct {
	APIEndpoint string
	ID          string
	Secret      string
}

type Client interface {
	CreateRecord(params CreateRecordParams) (Record, error)
	ReadRecord(id int) (Record, error)
	UpdateRecord(params UpdateRecordParams) (Record, error)
	DeleteRecord(id int) (Record, error)
	ListRecords(zone string) ([]Record, error)
}

type APIClient struct {
	httpClient  HTTPClient
	endpoint    string
	credentials Credentials
}

type Credentials struct {
	ClientID     string
	ClientSecret string
}

func NewClient(options ClientOptions) APIClient {
	return APIClient{
		httpClient: &http.Client{},
		endpoint:   options.APIEndpoint,
		credentials: Credentials{
			ClientID:     options.ID,
			ClientSecret: options.Secret,
		},
	}
}

type Record struct {
	ID        int
	Zone      string
	RR        dns.RR
	Comment   string
	CreatedAt time.Time
	UpdatedOn time.Time
}

type CreateRecordParams struct {
	Zone    string `json:"zone"`
	RR      string `json:"content"`
	Comment string `json:"comment"`
}

type UpdateRecordParams struct {
	ID      int    `json:"-"`
	Zone    string `json:"zone"`
	RR      string `json:"content"`
	Comment string `json:"comment"`
}

// Every field on the struct is available in the URL template.
// The struct is JSON marshalled into the request body.
func paramsToRequest(method string, url string, params any, credentials Credentials) (req *http.Request, err error) {
	tmpl, err := template.New("url").Parse(url)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, params)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err = http.NewRequest(method, b.String(), bytes.NewBuffer(body))
	req.SetBasicAuth(credentials.ClientID, credentials.ClientSecret)
	req.Header.Add("Content-Type", "application/json")

	paramsValue := reflect.ValueOf(params)
	query := req.URL.Query()
	for _, field := range reflect.VisibleFields(paramsValue.Type()) {
		if paramName := field.Tag.Get("param"); paramName != "" {
			fieldValue := paramsValue.FieldByName(field.Name)
			if fieldValue.Kind() != reflect.String {
				return nil, fmt.Errorf("param tagged field '%s' must be of type string", field.Name)
			}
			query.Set(paramName, fieldValue.String())
		}
	}
	req.URL.RawQuery = query.Encode()

	return
}

type APIError struct {
	parameterErr error
	status       int
	message      string
}

func (e *APIError) Error() string {
	if e.parameterErr == nil {
		return fmt.Sprintf("Status %d: %s", e.status, e.message)
	}
	return fmt.Sprintf("Status %d: %s: %v", e.status, e.message, e.parameterErr)
}

func (e *APIError) Unwrap() error {
	return e.parameterErr
}

type ParameterErrorField struct {
	Key     string
	Message string
}

type APIParameterError struct {
	FieldErrors []ParameterErrorField
	ParamErrors []ParameterErrorField
}

func (e *APIParameterError) Error() string {
	msgs := []string{}
	for _, fieldErr := range e.FieldErrors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", fieldErr.Key, fieldErr.Message))
	}
	return strings.Join(msgs, ", ") + "."
}

func parseErrorResponse(res *http.Response, params any) (error, error) {
	// Avoid reading the response body if successful.
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		return nil, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusBadRequest {
		paramErr := APIParameterError{
			FieldErrors: []ParameterErrorField{},
			ParamErrors: []ParameterErrorField{},
		}
		var parsedResponse rest.BadRequestErrorResponse
		if err = json.Unmarshal(body, &parsedResponse); err != nil {
			return nil, err
		}

		// Create mapping for field and param name conversion from API request body JSON
		// fields to struct fields.
		fieldNamesMap := map[string]string{}
		paramNamesMap := map[string]string{}
		for _, field := range reflect.VisibleFields(reflect.TypeOf(params)) {
			if fieldName := field.Tag.Get("json"); fieldName != "" {
				fieldNamesMap[fieldName] = field.Name
			}
			if paramName := field.Tag.Get("param"); paramName != "" {
				paramNamesMap[paramName] = field.Name
			}
		}

		// Map body JSON field errors to struct fields.
		for _, keyErr := range parsedResponse.Fields {
			paramErr.FieldErrors = append(paramErr.FieldErrors, ParameterErrorField{
				Key:     fieldNamesMap[keyErr.Key],
				Message: keyErr.Message,
			})

		}

		// Map param errors to struct fields.
		for _, keyErr := range parsedResponse.Params {
			paramErr.ParamErrors = append(paramErr.ParamErrors, ParameterErrorField{
				Key:     paramNamesMap[keyErr.Key],
				Message: keyErr.Message,
			})
		}

		return &APIError{
			message:      "bad parameters",
			status:       res.StatusCode,
			parameterErr: &paramErr,
		}, nil
	}

	// Other error types:
	var parsedResponse rest.ErrorResponse
	if err = json.Unmarshal(body, &parsedResponse); err != nil {
		return nil, err
	}

	return &APIError{
		message: parsedResponse.Message,
		status:  res.StatusCode,
	}, nil
}

func (c *APIClient) CreateRecord(params CreateRecordParams) (Record, error) {
	req, err := paramsToRequest(
		"POST",
		fmt.Sprintf("%s/records", c.endpoint),
		params,
		c.credentials,
	)
	if err != nil {
		return Record{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Record{}, err
	}
	apiErr, parsingErr := parseErrorResponse(res, params)
	if parsingErr != nil {
		return Record{}, parsingErr
	}
	if apiErr != nil {
		return Record{}, apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}

	var parsedRecord records.RecordResponse
	if err = json.Unmarshal(body, &parsedRecord); err != nil {
		return Record{}, err
	}

	rr, err := dns.NewRR(parsedRecord.Content)
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:        parsedRecord.ID,
		Zone:      parsedRecord.Zone,
		RR:        rr,
		Comment:   parsedRecord.Comment,
		CreatedAt: parsedRecord.CreatedAt,
		UpdatedOn: parsedRecord.UpdatedOn,
	}, nil
}

func (c *APIClient) ReadRecord(id int) (Record, error) {
	params := struct {
		ID int
	}{
		ID: id,
	}
	req, err := paramsToRequest(
		"GET",
		fmt.Sprintf("%s/records/{{ .ID }}", c.endpoint),
		params,
		c.credentials,
	)
	if err != nil {
		return Record{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Record{}, err
	}
	apiErr, parsingErr := parseErrorResponse(res, params)
	if parsingErr != nil {
		return Record{}, parsingErr
	}
	if apiErr != nil {
		return Record{}, apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}

	var parsedRecord records.RecordResponse
	if err = json.Unmarshal(body, &parsedRecord); err != nil {
		return Record{}, err
	}

	rr, err := dns.NewRR(parsedRecord.Content)
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:        parsedRecord.ID,
		Zone:      parsedRecord.Zone,
		RR:        rr,
		Comment:   parsedRecord.Comment,
		CreatedAt: parsedRecord.CreatedAt,
		UpdatedOn: parsedRecord.UpdatedOn,
	}, nil
}

func (c *APIClient) UpdateRecord(params UpdateRecordParams) (Record, error) {
	req, err := paramsToRequest(
		"PUT",
		fmt.Sprintf("%s/records/{{ .ID }}", c.endpoint),
		params,
		c.credentials,
	)
	if err != nil {
		return Record{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Record{}, err
	}
	apiErr, parsingErr := parseErrorResponse(res, params)
	if parsingErr != nil {
		return Record{}, parsingErr
	}
	if apiErr != nil {
		return Record{}, apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}

	var parsedRecord records.RecordResponse
	if err = json.Unmarshal(body, &parsedRecord); err != nil {
		return Record{}, err
	}

	rr, err := dns.NewRR(parsedRecord.Content)
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:        parsedRecord.ID,
		Zone:      parsedRecord.Zone,
		RR:        rr,
		Comment:   parsedRecord.Comment,
		CreatedAt: parsedRecord.CreatedAt,
		UpdatedOn: parsedRecord.UpdatedOn,
	}, nil
}

func (c *APIClient) DeleteRecord(id int) (Record, error) {
	params := struct {
		ID int
	}{
		ID: id,
	}
	req, err := paramsToRequest(
		"DELETE",
		fmt.Sprintf("%s/records/{{ .ID }}", c.endpoint),
		params,
		c.credentials,
	)
	if err != nil {
		return Record{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return Record{}, err
	}
	apiErr, parsingErr := parseErrorResponse(res, params)
	if parsingErr != nil {
		return Record{}, parsingErr
	}
	if apiErr != nil {
		return Record{}, apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}

	var parsedRecord records.RecordResponse
	if err = json.Unmarshal(body, &parsedRecord); err != nil {
		return Record{}, err
	}

	rr, err := dns.NewRR(parsedRecord.Content)
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:        parsedRecord.ID,
		Zone:      parsedRecord.Zone,
		RR:        rr,
		Comment:   parsedRecord.Comment,
		CreatedAt: parsedRecord.CreatedAt,
		UpdatedOn: parsedRecord.UpdatedOn,
	}, nil
}

func (c *APIClient) ListRecords(zone string) ([]Record, error) {
	params := struct {
		Zone string `param:"zone"`
	}{
		Zone: zone,
	}
	req, err := paramsToRequest(
		"GET",
		fmt.Sprintf("%s/records", c.endpoint),
		params,
		c.credentials,
	)
	if err != nil {
		return []Record{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return []Record{}, err
	}
	apiErr, parsingErr := parseErrorResponse(res, params)
	if parsingErr != nil {
		return []Record{}, parsingErr
	}
	if apiErr != nil {
		return []Record{}, apiErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []Record{}, err
	}

	var parsedRecords []records.RecordResponse
	if err = json.Unmarshal(body, &parsedRecords); err != nil {
		return []Record{}, err
	}

	records := make([]Record, len(parsedRecords))

	for i, record := range parsedRecords {
		rr, err := dns.NewRR(record.Content)
		if err != nil {
			return []Record{}, err
		}
		records[i] = Record{
			ID:        record.ID,
			Zone:      record.Zone,
			RR:        rr,
			Comment:   record.Comment,
			CreatedAt: record.CreatedAt,
			UpdatedOn: record.UpdatedOn,
		}
	}

	return records, nil
}
