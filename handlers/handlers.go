package handlers

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dpTopicApiModels "github.com/ONSdigital/dp-topic-api/models"
	dpTopicApiSdk "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"

	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	dpTopicApiSdkErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
)

// Constants...
const (
	dataEndpoint         = `\/data$`
	numOptsSummary       = 50
	maxMetadataOptions   = 1000
	maxAgeAndTimeOptions = 1000
	homepagePath         = "/"
	queryStrKey          = "showAll"
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
func getOptionsSummary(ctx context.Context, dc DatasetAPISdkClient, userAccessToken, collectionID, datasetID, edition, version string, dimensions dpDatasetApiSdk.VersionDimensionsList, numOpts int) (opts []dpDatasetApiSdk.VersionDimensionOptionsList, err error) {
	headers := dpDatasetApiSdk.Headers{
		CollectionID:    collectionID,
		UserAccessToken: userAccessToken,
	}

	for i := range dimensions.Items {
		dimension := &dimensions.Items[i]

		// for time and age, request all the options (assumed less than maxAgeAndTimeOptions)
		if dimension.Name == mapper.DimensionTime || dimension.Name == mapper.DimensionAge {
			// query with limit maxAgeAndTimeOptions
			q := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: maxAgeAndTimeOptions}
			opt, err := dc.GetVersionDimensionOptions(ctx, headers, datasetID, edition, version, dimension.Name, &q)
			if err != nil {
				return opts, err
			}

			totalCount := len(opt.Items)

			if totalCount > maxAgeAndTimeOptions {
				log.Warn(ctx, "total number of options is greater than the requested number", log.Data{"max_age_and_time_options": maxAgeAndTimeOptions, "total_count": totalCount})
			}

			opts = append(opts, opt)
			continue
		}

		// for other dimensions, cap the number of options to numOpts
		q := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: numOpts}
		opt, err := dc.GetVersionDimensionOptions(ctx, headers, datasetID, edition, version, dimension.Name, &q)
		if err != nil {
			return opts, err
		}
		opts = append(opts, opt)
	}
	return opts, nil
}

// getText gets a byte array containing the metadata content, based on options returned by dataset API.
// If a dimension has more than maxMetadataOptions, an error will be returned
func getText(ctx context.Context, dc DatasetAPISdkClient, headers dpDatasetApiSdk.Headers, datasetID, editionID, versionID string,
	metadata dpDatasetApiModels.Metadata, dimensions dpDatasetApiSdk.VersionDimensionsList) ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(metadata.ToString())
	b.WriteString("Dimensions:\n")

	for i := range dimensions.Items {
		dimension := &dimensions.Items[i]
		q := dpDatasetApiSdk.QueryParams{Offset: 0, Limit: maxMetadataOptions}
		options, err := dc.GetVersionDimensionOptions(ctx, headers, datasetID, editionID, versionID, dimension.Name, &q)
		if err != nil {
			return nil, err
		}
		if len(options.Items) > maxMetadataOptions {
			return []byte{}, errTooManyOptions
		}

		b.WriteString(options.ToString())
	}

	return b.Bytes(), nil
}

func handleRequestForZebedeeJSONData(ctx context.Context, w http.ResponseWriter, zc ZebedeeClient, path, userAccessToken string) (wasZebedeeRequest bool) {
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

// sorts options by code - numerically if possible, with negatives listed last
func sortOptionsByCode(items []dpDatasetApiModels.PublicDimensionOption) []dpDatasetApiModels.PublicDimensionOption {
	sorted := []dpDatasetApiModels.PublicDimensionOption{}
	sorted = append(sorted, items...)

	doNumericSort := func(items []dpDatasetApiModels.PublicDimensionOption) bool {
		for i := range items {
			item := &items[i]
			_, err := strconv.Atoi(item.Option)
			if err != nil {
				return false
			}
		}
		return true
	}

	if doNumericSort(items) {
		sort.Slice(sorted, func(i, j int) bool {
			left, _ := strconv.Atoi(sorted[i].Option)
			right, _ := strconv.Atoi(sorted[j].Option)
			if left*right < 0 {
				return right < 0
			} else {
				return left*left < right*right
			}
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Option < sorted[j].Option
		})
	}
	return sorted
}

func mapDimensionCategories(dimCategories population.GetDimensionCategoriesResponse) map[string]population.DimensionCategory {
	dimensionCategoryMap := make(map[string]population.DimensionCategory)
	for _, dimensionCategory := range dimCategories.Categories {
		dimensionCategoryMap[dimensionCategory.Id] = dimensionCategory
	}
	return dimensionCategoryMap
}

// sorts population.DimensionCategoryItems - numerically if possible, with negatives listed last
func sortCategoriesByID(items []population.DimensionCategoryItem) []population.DimensionCategoryItem {
	sorted := []population.DimensionCategoryItem{}
	sorted = append(sorted, items...)

	doNumericSort := func(items []population.DimensionCategoryItem) bool {
		for _, item := range items {
			_, err := strconv.Atoi(item.ID)
			if err != nil {
				return false
			}
		}
		return true
	}

	if doNumericSort(items) {
		sort.Slice(sorted, func(i, j int) bool {
			left, _ := strconv.Atoi(sorted[i].ID)
			right, _ := strconv.Atoi(sorted[j].ID)
			if left*right < 0 {
				return right < 0
			} else {
				return left*left < right*right
			}
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].ID < sorted[j].ID
		})
	}
	return sorted
}

func GetPublicOrPrivateTopics(
	topicsClient TopicAPIClient,
	cfg config.Config,
	ctx context.Context,
	topicHeaders dpTopicApiSdk.Headers,
	topicID string,
) (*dpTopicApiModels.Topic, dpTopicApiSdkErrors.Error) {
	if cfg.IsPublishing {
		topicResponse, err := topicsClient.GetTopicPrivate(ctx, topicHeaders, topicID)
		if err != nil {
			return nil, err
		} else {
			return topicResponse.Current, err
		}
	} else {
		return topicsClient.GetTopicPublic(ctx, topicHeaders, topicID)
	}
}
