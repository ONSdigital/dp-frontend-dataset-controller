package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/ONSdigital/go-ns/log"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/pkg/errors"
)

const service = "dataset-api"

// ErrInvalidDatasetAPIResponse is returned when the dataset api does not respond
// with a valid status
type ErrInvalidDatasetAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidDatasetAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from dataset api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from dataset api if an error is returned
func (e ErrInvalidDatasetAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidDatasetAPIResponse{}

// Client is a dataset api client which can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
}

// closeResponseBody closes the response body and logs an error containing the context if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.ErrorCtx(ctx, err, log.Data{"message": "error closing http response body"})
	}
}

// NewAPIClient creates a new instance of Client with a given dataset api url and the relevant tokens
func NewAPIClient(datasetAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: datasetAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()

	resp, err := c.cli.Get(ctx, c.url+"/healthcheck")
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return service, NewDatasetAPIResponse(resp, "/healthcheck")
	}

	return service, nil
}

// Get returns dataset level information for a given dataset id
func (c *Client) Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m Model, err error) {
	uri := fmt.Sprintf("%s/datasets/%s", c.url, datasetID)

	clientlog.Do(ctx, "retrieving dataset", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && (common.IsCallerPresent(ctx) || common.IsFlorenceIdentityPresent(ctx)) {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetByPath returns dataset level information for a given dataset path
func (c *Client) GetByPath(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, path string) (m Model, err error) {
	uri := fmt.Sprintf("%s/%s", c.url, strings.Trim(path, "/"))

	clientlog.Do(ctx, "retrieving data from dataset API", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && (common.IsCallerPresent(ctx) || common.IsFlorenceIdentityPresent(ctx)) {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetDatasets returns the list of datasets
func (c *Client) GetDatasets(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string) (m ModelCollection, err error) {
	uri := fmt.Sprintf("%s/datasets", c.url)

	clientlog.Do(ctx, "retrieving datasets", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	return
}

// GetEdition retrieves a single edition document from a given datasetID and edition label
func (c *Client) GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (m Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s", c.url, datasetID, edition)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	if next, ok := body["next"]; ok && common.IsCallerPresent(ctx) {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetEditions returns all editions for a dataset
func (c *Client) GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions", c.url, datasetID)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return nil, nil
	}

	if _, ok := body["items"].([]interface{})[0].(map[string]interface{})["next"]; ok && common.IsCallerPresent(ctx) {
		var items []map[string]interface{}
		for _, item := range body["items"].([]interface{}) {
			items = append(items, item.(map[string]interface{})["next"].(map[string]interface{}))
		}
		parentItems := make(map[string]interface{})
		parentItems["items"] = items
		b, err = json.Marshal(parentItems)
		if err != nil {
			return
		}
	}

	editions := struct {
		Items []Edition `json:"items"`
	}{}
	err = json.Unmarshal(b, &editions)
	m = editions.Items
	return
}

// GetVersions gets all versions for an edition from the dataset api
func (c *Client) GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string) (m []Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions", c.url, datasetID, edition)

	clientlog.Do(ctx, "retrieving dataset versions", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	versions := struct {
		Items []Version `json:"items"`
	}{}

	err = json.Unmarshal(b, &versions)
	m = versions.Items
	return
}

// GetVersion gets a specific version for an edition from the dataset api
func (c *Client) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, datasetID, edition, version)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetInstance returns an instance from the dataset api
func (c *Client) GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) (m Instance, err error) {
	uri := fmt.Sprintf("%s/instances/%s", c.url, instanceID)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// PutVersion update the version
func (c *Client) PutVersion(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, v Version) error {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, datasetID, edition, version)
	clientlog.Do(ctx, "updating version", service, uri)

	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "error while attempting to create http request")
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("incorrect http status, expected: 200, actual: %d, uri: %s", resp.StatusCode, uri)
	}
	return nil
}

// GetMetadataURL returns the URL for the metadata of a given dataset id, edition and version
func (c *Client) GetMetadataURL(id, edition, version string) string {
	return fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/metadata", c.url, id, edition, version)
}

// GetVersionMetadata returns the metadata for a given dataset id, edition and version
func (c *Client) GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m Metadata, err error) {
	uri := c.GetMetadataURL(id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version metadata", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetDimensions will return a versions dimensions
func (c *Client) GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m Dimensions, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions", c.url, id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version dimensions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	sort.Sort(m.Items)

	return
}

// GetOptions will return the options for a dimension
func (c *Client) GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string) (m Options, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options", c.url, id, edition, version, dimension)

	clientlog.Do(ctx, "retrieving options for dimension", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// NewDatasetAPIResponse creates an error response, optionally adding body to e when status is 404
func NewDatasetAPIResponse(resp *http.Response, uri string) (e *ErrInvalidDatasetAPIResponse) {
	e = &ErrInvalidDatasetAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read DatasetAPI body"
			return
		}
		defer closeResponseBody(nil, resp)

		e.body = string(b)
	}
	return
}

func addCollectionIDHeader(r *http.Request, collectionID string) {
	if len(collectionID) > 0 {
		r.Header.Add(common.CollectionIDHeaderKey, collectionID)
	}
}

// doGetWithAuthHeaders executes clienter.Do setting the user and service authentication token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	return c.cli.Do(ctx, req)
}

// doGetWithAuthHeadersAndWithDownloadToken executes clienter.Do setting the user and service authentication and dwonload token token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeadersAndWithDownloadToken(ctx context.Context, userAuthToken, serviceAuthToken, downloadserviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	common.AddDownloadServiceTokenHeader(req, downloadserviceAuthToken)
	return c.cli.Do(ctx, req)
}
