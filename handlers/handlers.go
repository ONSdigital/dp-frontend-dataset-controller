package handlers

import (
	"bytes"
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"net/http"
	"regexp"
)

// Constants...
const (
	dataEndpoint         = `\/data$`
	numOptsSummary       = 50
	maxMetadataOptions   = 1000
	maxAgeAndTimeOptions = 1000
	homepagePath         = "/"
)

var errTooManyOptions = errors.New("too many options in dimension")

func setStatusCode(ctx context.Context, w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if err == errTooManyOptions {
		status = http.StatusRequestEntityTooLarge
	}
	if err, ok := err.(ClientError); ok {
		if err.Code() == http.StatusNotFound {
			status = err.Code()
		}
	}
	log.Error(ctx, "client error", err, log.Data{"setting-response-status": status})
	w.WriteHeader(status)
}

// getOptionsSummary requests a maximum of numOpts for each dimension, and returns the array of Options structs for each dimension, each one containing up to numOpts options.
func getOptionsSummary(ctx context.Context, dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, dimensions dataset.VersionDimensions, numOpts int) (opts []dataset.Options, err error) {
	for _, dim := range dimensions.Items {

		// for time and age, request all the options (assumed less than maxAgeAndTimeOptions)
		if dim.Name == mapper.DimensionTime || dim.Name == mapper.DimensionAge {

			// query with limit maxAgeAndTimeOptions
			q := dataset.QueryParams{Offset: 0, Limit: maxAgeAndTimeOptions}
			opt, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
			if err != nil {
				return opts, err
			}

			if opt.TotalCount > maxAgeAndTimeOptions {
				log.Warn(ctx, "total number of options is greater than the requested number", log.Data{"max_age_and_time_options": maxAgeAndTimeOptions, "total_count": opt.TotalCount})
			}

			opts = append(opts, opt)
			continue
		}

		// for other dimensions, cap the number of options to numOpts
		q := dataset.QueryParams{Offset: 0, Limit: numOpts}
		opt, err := dc.GetOptions(ctx, userAccessToken, "", collectionID, datasetID, edition, version, dim.Name, &q)
		if err != nil {
			return opts, err
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

// getText gets a byte array containing the metadata content, based on options returned by dataset API.
// If a dimension has more than maxMetadataOptions, an error will be returned
func getText(dc DatasetClient, userAccessToken, collectionID, datasetID, edition, version string, metadata dataset.Metadata, dimensions dataset.VersionDimensions, req *http.Request) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")

	for _, dimension := range dimensions.Items {
		q := dataset.QueryParams{Offset: 0, Limit: maxMetadataOptions}
		options, err := dc.GetOptions(req.Context(), userAccessToken, "", collectionID, datasetID, edition, version, dimension.Name, &q)
		if err != nil {
			return nil, err
		}
		if options.TotalCount > maxMetadataOptions {
			return []byte{}, errTooManyOptions
		}

		b.WriteString(options.String())
	}

	return b.Bytes(), nil
}

func handleRequestForZebedeeJsonData(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, path string, ctx context.Context, userAccessToken string) (wasZebedeeRequest bool) {
	wasZebedeeRequest = false
	// Since MatchString will only error if the regex is invalid, and the regex is
	// constant, don't capture the error
	if ok, _ := regexp.MatchString(dataEndpoint, path); ok {
		wasZebedeeRequest = true
		strippedPath := path[0:(len(path) - 5)] // i.e. remove the "/data" at the end of the path

		b, err := zc.Get(ctx, userAccessToken, "/data?uri="+strippedPath)
		if err != nil {
			setStatusCode(ctx, w, err)
			return
		}
		_, err = w.Write(b)
		if err != nil {
			setStatusCode(ctx, w, errors.Wrap(err, "failed to write zebedee client get response"))
		}
	}

	return
}
