package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	cli *http.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		cli: &http.Client{Timeout: 5 * time.Second},
		url: filterAPIURL,
	}
}

//FilterJob represents a filter job response from the filter api
type FilterJob struct {
	DatasetFilterID string `json:"dataset_filter_id"`
	FilterJobID     string `json:"filter_job_id"`
	State           string `json:"state"`
}

// CreateJob creates a filter job and returns the associated filterJobID
func (c *Client) CreateJob(datasetFilterID string) (string, error) {
	fj := FilterJob{DatasetFilterID: "384900328", State: "created"} // DatasetFilterId comes from dataset api

	b, err := json.Marshal(fj)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.url+"/filters", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("invalid status from filter api")
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &fj); err != nil {
		return "", err
	}

	return fj.FilterJobID, nil
}

func (c *Client) AddDimension(id, name string) error {
	resp, err := http.Post(fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, id, name), "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}
