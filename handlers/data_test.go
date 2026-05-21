package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/clients"
	topicAPIModels "github.com/ONSdigital/dp-topic-api/models"
	topicAPISDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicAPISDKErrors "github.com/ONSdigital/dp-topic-api/sdk/errors"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testDatasetHeaders = datasetAPISDK.Headers{AccessToken: testUserAccessToken}
	testTopicHeaders   = topicAPISDK.Headers{UserAuthToken: testUserAccessToken}

	testTopicIDs       = []string{"topic-economy-id", "topic-inflation-id"}
	testTopicSlugs     = []string{"economy", "inflation"}
	testTopicEconomy   = &topicAPIModels.Topic{ID: "topic-economy-id", Slug: "economy"}
	testTopicInflation = &topicAPIModels.Topic{ID: "topic-inflation-id", Slug: "inflation"}

	testStaticDataset = datasetAPIModels.Dataset{
		ID:          "dataset-123",
		Title:       "Producer price inflation (MM22)",
		Description: "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
		Keywords:    []string{"manufacturing", "input prices", "output prices", "producer prices"},
		NextRelease: "To be announced",
		Topics:      testTopicIDs,
		Contacts: []datasetAPIModels.ContactDetails{
			{
				Name:      "Business Prices team",
				Email:     "business.prices@ons.gov.uk",
				Telephone: "+44 1633 456907",
			},
		},
		QMI: &datasetAPIModels.GeneralDetails{
			Title:       "Producer Prices QMI",
			Description: "Quality and Methodology Information for Producer Price Indices",
			HRef:        "https://www.ons.gov.uk/economy/inflationandpriceindices/qmis/producerpriceindicesqmi",
		},
		Links: &datasetAPIModels.DatasetLinks{
			LatestVersion: &datasetAPIModels.LinkObject{ID: "3"},
		},
		Type: datasetAPIModels.Static.String(),
	}

	testStaticVersion = datasetAPIModels.Version{
		Version:            3,
		Edition:            "2025",
		EditionTitle:       "2025 edition",
		ReleaseDate:        "2025-01-15T00:00:00.000Z",
		QualityDesignation: datasetAPIModels.QualityDesignationAccreditedOfficial,
		Distributions: &[]datasetAPIModels.Distribution{
			{
				Title:       "file.csv",
				DownloadURL: "http://localhost:23600/downloads/file.csv",
			},
			{
				Title:       "file.xls",
				DownloadURL: "http://localhost:23600/downloads/file.xls",
			},
		},
	}

	testStaticPreviousVersions = []datasetAPIModels.Version{
		{
			Version:     2,
			Edition:     "2024",
			ReleaseDate: "2024-01-15T00:00:00.000Z",
			Alerts: &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeCorrection, Description: "Correction for version 2"},
			},
		},
		{
			Version:     1,
			Edition:     "2023",
			ReleaseDate: "2023-01-15T00:00:00.000Z",
			Alerts: &[]datasetAPIModels.Alert{
				{Type: datasetAPIModels.AlertTypeAlert, Description: "Alert for version 1"},
			},
		},
	}

	testFullVersionsList = datasetAPISDK.VersionsList{
		Items: []datasetAPIModels.Version{
			testStaticVersion,
			testStaticPreviousVersions[0],
			testStaticPreviousVersions[1],
		},
		Count:      3,
		TotalCount: 3,
	}
)

func TestDatasetData(t *testing.T) {
	ctx := gomock.Any()

	dataset := testStaticDataset
	datasetID := dataset.ID
	requestPath := fmt.Sprintf("/%s/datasets/%s/data", testTopicSlugs[0], datasetID)
	urlVars := map[string]string{"topic": testTopicSlugs[0], "datasetID": datasetID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := clients.NewMockDatasetAPISdkClient(ctrl)
	mockTopicClient := clients.NewMockTopicAPIClient(ctrl)

	Convey("Given datasetData handler", t, func() {
		Convey("When dataset is static and topic matches", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 200 with the expected JSON body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				expectedResponseBody := &zebedee.DatasetLandingPage{
					Type: zebedee.PageTypeDatasetLandingPage,
					URI:  "/economy/datasets/dataset-123",
					Description: zebedee.Description{
						DatasetID:       "dataset-123",
						Title:           "Producer price inflation (MM22)",
						Summary:         "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						MetaDescription: "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						Keywords:        []string{"manufacturing", "input prices", "output prices", "producer prices"},
						NextRelease:     "To be announced",
						CanonicalTopic:  testTopicSlugs[0],
						Topics:          []string{testTopicSlugs[1]},
						Contact: zebedee.Contact{
							Name:      "Business Prices team",
							Email:     "business.prices@ons.gov.uk",
							Telephone: "+44 1633 456907",
						},
					},
					RelatedMethodology: []zebedee.Link{
						{
							Title:   "Producer Prices QMI",
							Summary: "Quality and Methodology Information for Producer Price Indices",
							URI:     "/economy/inflationandpriceindices/qmis/producerpriceindicesqmi",
						},
					},
				}

				var resp zebedee.DatasetLandingPage
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(&resp, ShouldResemble, expectedResponseBody)
			})
		})

		Convey("When GetDataset fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{}, errors.New("failed to fetch dataset"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When dataset type is not static", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{ID: datasetID, Type: "filterable"}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When topic client fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("topic API error")})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When canonical topic does not match topic slug in URL", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(&topicAPIModels.Topic{ID: "another-topic-id", Slug: "different-topic"}, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When mapper fails due to invalid QMI URL", func() {
			datasetWithInvalidQMI := dataset
			datasetWithInvalidQMI.QMI = &datasetAPIModels.GeneralDetails{
				Title:       "QMI",
				Description: "Invalid QMI URL",
				HRef:        "https://[invalid:url",
			}

			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetWithInvalidQMI, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			datasetData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestEditionData(t *testing.T) {
	ctx := gomock.Any()

	dataset := testStaticDataset
	datasetID := dataset.ID
	editionID := testStaticVersion.Edition
	requestPath := fmt.Sprintf("/%s/datasets/%s/editions/%s/data", testTopicSlugs[0], datasetID, editionID)
	urlVars := map[string]string{"topic": testTopicSlugs[0], "datasetID": datasetID, "editionID": editionID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := clients.NewMockDatasetAPISdkClient(ctrl)
	mockTopicClient := clients.NewMockTopicAPIClient(ctrl)

	Convey("Given editionData handler", t, func() {
		Convey("When dataset and edition are static and topic matches", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
				Return(testFullVersionsList, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 200 with the expected JSON body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				expectedResponseBody := &zebedee.Dataset{
					Description: zebedee.Description{
						DatasetID:         "dataset-123",
						Title:             "Producer price inflation (MM22)",
						Summary:           "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						MetaDescription:   "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						Contact:           zebedee.Contact{Name: "Business Prices team", Email: "business.prices@ons.gov.uk", Telephone: "+44 1633 456907"},
						Keywords:          []string{"manufacturing", "input prices", "output prices", "producer prices"},
						ReleaseDate:       "2025-01-15T00:00:00.000Z",
						NextRelease:       "To be announced",
						CanonicalTopic:    "economy",
						Topics:            []string{"inflation"},
						Edition:           "2025 edition",
						NationalStatistic: true,
					},
					Type: zebedee.PageTypeDataset,
					Downloads: []zebedee.Download{
						{
							File: "file.csv",
							URI:  "http://localhost:23600/downloads/file.csv",
						},
						{
							File: "file.xls",
							URI:  "http://localhost:23600/downloads/file.xls",
						},
					},
					URI: "/economy/datasets/dataset-123/editions/2025",
					Versions: []zebedee.Version{
						{
							URI:         "/economy/datasets/dataset-123/editions/2024/versions/2",
							ReleaseDate: "2024-01-15T00:00:00.000Z",
							Notice:      "Correction for version 2",
						},
						{
							URI:         "/economy/datasets/dataset-123/editions/2023/versions/1",
							ReleaseDate: "2023-01-15T00:00:00.000Z",
							Notice:      "",
						},
					},
				}

				var resp zebedee.Dataset
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(&resp, ShouldResemble, expectedResponseBody)
			})
		})

		Convey("When GetDataset fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{}, errors.New("failed to fetch dataset"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When dataset type is not static", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{ID: datasetID, Type: "filterable"}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When topic client fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("topic API error")})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When canonical topic does not match topic slug in URL", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(&topicAPIModels.Topic{ID: "another-topic-id", Slug: "different-topic"}, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When GetVersions fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
				Return(datasetAPISDK.VersionsList{}, errors.New("failed to fetch versions"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When mapper fails due to empty dataset ID", func() {
			emptyIDDataset := dataset
			emptyIDDataset.ID = ""

			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(emptyIDDataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: 1000}).
				Return(testFullVersionsList, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			editionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestVersionData(t *testing.T) {
	ctx := gomock.Any()

	dataset := testStaticDataset
	datasetID := dataset.ID
	editionID := testStaticVersion.Edition
	versionID := fmt.Sprintf("%d", testStaticVersion.Version)
	requestPath := fmt.Sprintf("/%s/datasets/%s/editions/%s/versions/%s/data", testTopicSlugs[0], datasetID, editionID, versionID)
	urlVars := map[string]string{
		"topic":     testTopicSlugs[0],
		"datasetID": datasetID,
		"editionID": editionID,
		"versionID": versionID,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDatasetClient := clients.NewMockDatasetAPISdkClient(ctrl)
	mockTopicClient := clients.NewMockTopicAPIClient(ctrl)

	Convey("Given versionData handler", t, func() {
		Convey("When dataset and version are static and topic matches", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersionV2(ctx, testDatasetHeaders, datasetID, editionID, versionID).
				Return(testStaticVersion, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: testStaticVersion.Version - 1, Offset: 1}).
				Return(datasetAPISDK.VersionsList{Items: testStaticPreviousVersions}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 200 with the expected JSON body", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				expectedResponseBody := &zebedee.Dataset{
					Description: zebedee.Description{
						DatasetID:         "dataset-123",
						Title:             "Producer price inflation (MM22)",
						Summary:           "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						MetaDescription:   "UK price index data at manufacturing, aggregated industry and product group levels. Data supplied from individual manufacturers, importers and exporters. Monthly and annual data.",
						Contact:           zebedee.Contact{Name: "Business Prices team", Email: "business.prices@ons.gov.uk", Telephone: "+44 1633 456907"},
						Keywords:          []string{"manufacturing", "input prices", "output prices", "producer prices"},
						ReleaseDate:       "2025-01-15T00:00:00.000Z",
						NextRelease:       "To be announced",
						CanonicalTopic:    "economy",
						Topics:            []string{"inflation"},
						Edition:           "2025 edition",
						NationalStatistic: true,
					},
					Type: zebedee.PageTypeDataset,
					Downloads: []zebedee.Download{
						{
							File: "file.csv",
							URI:  "http://localhost:23600/downloads/file.csv",
						},
						{
							File: "file.xls",
							URI:  "http://localhost:23600/downloads/file.xls",
						},
					},
					URI: "/economy/datasets/dataset-123/editions/2025/versions/3",
					Versions: []zebedee.Version{
						{
							URI:         "/economy/datasets/dataset-123/editions/2024/versions/2",
							ReleaseDate: "2024-01-15T00:00:00.000Z",
							Notice:      "Correction for version 2",
						},
						{
							URI:         "/economy/datasets/dataset-123/editions/2023/versions/1",
							ReleaseDate: "2023-01-15T00:00:00.000Z",
							Notice:      "",
						},
					},
				}

				var resp zebedee.Dataset
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				So(err, ShouldBeNil)
				So(&resp, ShouldResemble, expectedResponseBody)
			})
		})

		Convey("When GetDataset fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{}, errors.New("failed to fetch dataset"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When dataset type is not static", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(datasetAPIModels.Dataset{ID: datasetID, Type: "filterable"}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When topic client fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(nil, topicAPISDKErrors.StatusError{Code: http.StatusInternalServerError, Err: errors.New("topic API error")})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When canonical topic does not match topic slug in URL", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(&topicAPIModels.Topic{ID: "another-topic-id", Slug: "different-topic"}, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 404 Not Found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})

		Convey("When GetVersionV2 fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersionV2(ctx, testDatasetHeaders, datasetID, editionID, versionID).
				Return(datasetAPIModels.Version{}, errors.New("failed to fetch version"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When previous versions fetch fails", func() {
			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(dataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersionV2(ctx, testDatasetHeaders, datasetID, editionID, versionID).
				Return(testStaticVersion, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: testStaticVersion.Version - 1, Offset: 1}).
				Return(datasetAPISDK.VersionsList{}, errors.New("failed to fetch previous versions"))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When mapper fails due to empty dataset ID", func() {
			emptyIDDataset := dataset
			emptyIDDataset.ID = ""

			mockDatasetClient.EXPECT().GetDataset(ctx, testDatasetHeaders, datasetID).
				Return(emptyIDDataset, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[0]).
				Return(testTopicEconomy, nil)

			mockTopicClient.EXPECT().GetTopicPublic(ctx, testTopicHeaders, dataset.Topics[1]).
				Return(testTopicInflation, nil)

			mockDatasetClient.EXPECT().GetVersionV2(ctx, testDatasetHeaders, datasetID, editionID, versionID).
				Return(testStaticVersion, nil)

			mockDatasetClient.EXPECT().GetVersions(ctx, testDatasetHeaders, datasetID, editionID, &datasetAPISDK.QueryParams{Limit: testStaticVersion.Version - 1, Offset: 1}).
				Return(datasetAPISDK.VersionsList{Items: testStaticPreviousVersions}, nil)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, requestPath, http.NoBody)
			r = mux.SetURLVars(r, urlVars)

			versionData(r, w, mockDatasetClient, mockTopicClient, false, testUserAccessToken)

			Convey("Then the response status code should be 500 Internal Server Error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}
