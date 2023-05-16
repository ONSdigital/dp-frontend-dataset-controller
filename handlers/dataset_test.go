package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/assets"
	"github.com/ONSdigital/dp-frontend-dataset-controller/cache"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	dsp "github.com/ONSdigital/dp-frontend-dataset-controller/model/dataset"
	render "github.com/ONSdigital/dp-renderer/v2"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	datasetLandingPageURI = "/peoplepopulationandcommunity/birthsdeathsandmarriages/deaths/datasets/weeklyprovisionalfiguresondeathsregisteredinenglandandwales"
	editionPageURI        = datasetLandingPageURI + "/2022"
)

const (
	userAuthTokenDatasets       = "123456789"
	collectionIDDatasets        = "testing-collection-123456789"
	localeDatasets              = "cy"
	staticFilesDownloadEndpoint = "downloads-new"
)

func TestDatasetHandlers(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	mockZebedeeClient := NewMockZebedeeClient(mockCtrl)
	mockFilesAPIClient := NewMockFilesAPIClient(mockCtrl)
	mockRend := NewMockRenderClient(mockCtrl)

	Convey("DatasetPage handler with non /data endpoint", t, func() {
		expectedDownloadFilename := "download_filename.csv"
		expectedVersionURI := editionPageURI + "/previous/v1"

		cfg := initialiseMockConfig()
		hp := zebedee.HomepageContent{}
		bc := []zebedee.Breadcrumb{{URI: editionPageURI}, {URI: datasetLandingPageURI}}

		Convey("Given file data stored in Files API", func() {
			dlp := zebedee.DatasetLandingPage{
				URI:      datasetLandingPageURI,
				Datasets: []zebedee.Link{{URI: editionPageURI}},
			}
			editionDataSet := zebedee.Dataset{
				URI:      editionPageURI,
				Versions: []zebedee.Version{{URI: expectedVersionURI}},
			}
			versionDataSet := zebedee.Dataset{
				URI:       expectedVersionURI,
				Downloads: []zebedee.Download{{URI: expectedDownloadFilename}},
			}

			mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

			mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, homepagePath).Return(hp, nil)
			mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, datasetLandingPageURI).Return(dlp, nil)
			mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionDataSet.URI).Return(bc, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionPageURI).Return(editionDataSet, nil)
			mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, expectedVersionURI).Return(versionDataSet, nil)

			var actualPageModel mapper.DatasetPage

			mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset").Do(func(w io.Writer, pageModel interface{}, templateName string) {
				actualPageModel = pageModel.(mapper.DatasetPage)
			})

			Convey("And that retrieving file data from files api is successful", func() {
				expectedDownloadFileSize := "100"
				expectedDownloadFileSizeInt, _ := strconv.Atoi(expectedDownloadFileSize)

				fmd := files.FileMetaData{SizeInBytes: uint64(expectedDownloadFileSizeInt)}

				mockFilesAPIClient.EXPECT().GetFile(ctx, expectedDownloadFilename, userAuthTokenDatasets).Return(fmd, nil)

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then the request to generate is OK", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})

					Convey("And the download size is known", func() {
						actualDownloadSize := actualPageModel.DatasetPage.Versions[0].Downloads[0].Size
						So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
					})
				})
			})

			Convey("And that retrieving file data from Files API is unsuccessful", func() {
				mockFilesAPIClient.EXPECT().GetFile(ctx, expectedDownloadFilename, userAuthTokenDatasets).Return(files.FileMetaData{}, errors.New("files api broken"))

				w, req := generateRecorderAndRequest()

				ctxOther := context.Background()
				mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
				So(err, ShouldBeNil)

				DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

				Convey("When the dataset page is rendered", func() {
					actualDownloadSize := actualPageModel.DatasetPage.Versions[0].Downloads[0].Size

					Convey("Then the request to generate is OK", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})

					Convey("And the file size is 0 (unknown)", func() {
						So(actualDownloadSize, ShouldEqual, "0")
					})
				})
			})
		})

		Convey("Given file data stored in Zebedee", func() {
			expectedDownloadFileSize := "100"

			dlp := zebedee.DatasetLandingPage{
				URI:      datasetLandingPageURI,
				Datasets: []zebedee.Link{{URI: editionPageURI}},
			}
			editionDataSet := zebedee.Dataset{
				URI:      editionPageURI,
				Versions: []zebedee.Version{{URI: expectedVersionURI}},
			}
			versionDataSet := zebedee.Dataset{
				URI:       expectedVersionURI,
				Downloads: []zebedee.Download{{File: expectedDownloadFilename, Size: expectedDownloadFileSize}},
			}

			Convey("And all page data is successfully retrieved", func() {
				mockRend.EXPECT().NewBasePageModel().Return(coreModel.NewPage(cfg.PatternLibraryAssetsPath, cfg.SiteDomain))

				mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, homepagePath).Return(hp, nil)
				mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, datasetLandingPageURI).Return(dlp, nil)
				mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionDataSet.URI).Return(bc, nil)
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionPageURI).Return(editionDataSet, nil)
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, expectedVersionURI).Return(versionDataSet, nil)

				var actualPageModel mapper.DatasetPage

				mockRend.EXPECT().BuildPage(gomock.Any(), gomock.Any(), "dataset").Do(func(w io.Writer, pageModel interface{}, templateName string) {
					actualPageModel = pageModel.(mapper.DatasetPage)
				})

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)
					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then the request to generate is OK", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
					})

					Convey("And the file size is known", func() {
						actualDownloadSize := actualPageModel.DatasetPage.Versions[0].Downloads[0].Size
						So(actualDownloadSize, ShouldEqual, expectedDownloadFileSize)
					})
				})
			})

			Convey("Given retrieving a versions dataset metadata is unsuccessful", func() {
				mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, homepagePath).Return(hp, nil)
				mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, datasetLandingPageURI).Return(dlp, nil)
				mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionDataSet.URI).Return(bc, nil)
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, editionPageURI).Return(editionDataSet, nil)
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, expectedVersionURI).Return(zebedee.Dataset{}, errors.New("Error retrieving version metadata"))

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then an internal server error is the response code", func() {
						So(w.Code, ShouldEqual, http.StatusInternalServerError)
					})
				})
			})

			Convey("And retrieving the main dataset's metadata is unsuccessful", func() {
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).
					Return(zebedee.Dataset{}, errors.New("something went wrong :("))

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then an internal server error is the response code", func() {
						So(w.Code, ShouldEqual, http.StatusInternalServerError)
					})
				})
			})

			Convey("And retrieving the breadcrumb is unsuccessful", func() {
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(zebedee.Dataset{}, nil)
				mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(nil, errors.New("something went wrong"))

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then an internal server error is the response code", func() {
						So(w.Code, ShouldEqual, http.StatusInternalServerError)
					})
				})
			})

			Convey("And the breadcrumb is too short", func() {
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(zebedee.Dataset{}, nil)
				mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return([]zebedee.Breadcrumb{{URI: "TOO/SHORT"}}, nil)

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then an internal server error is the response code", func() {
						So(w.Code, ShouldEqual, http.StatusInternalServerError)
					})
				})
			})

			Convey("And retrieving the Dataset landing page is unsuccessful", func() {
				mockZebedeeClient.EXPECT().GetHomepageContent(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(hp, nil)
				mockZebedeeClient.EXPECT().GetDatasetLandingPage(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(zebedee.DatasetLandingPage{}, errors.New("something went wrong :("))
				mockZebedeeClient.EXPECT().GetBreadcrumb(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(bc, nil)
				mockZebedeeClient.EXPECT().GetDataset(ctx, userAuthTokenDatasets, collectionIDDatasets, localeDatasets, gomock.Any()).Return(zebedee.Dataset{}, nil)

				Convey("When the dataset page is rendered", func() {
					w, req := generateRecorderAndRequest()

					ctxOther := context.Background()
					mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
					So(err, ShouldBeNil)

					DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

					Convey("Then an internal server error is the response code", func() {
						So(w.Code, ShouldEqual, http.StatusInternalServerError)
					})
				})
			})
		})
	})

	Convey("Given a request for an endpoint ending in /data", t, func() {
		Convey("When path contains /data", func() {
			path := "/path/to/some"
			zebedeePath := "/data?uri=" + path
			expectedBody := []byte("some content")

			w, req := generateRecorderAndRequest()
			req.URL.Path = path + "/data"

			mockZebedeeClient.EXPECT().Get(ctx, userAuthTokenDatasets, zebedeePath).Return(expectedBody, nil)

			cfg, err := config.Get()
			So(err, ShouldBeNil)

			ctxOther := context.Background()
			mockCacheList, err := cache.GetMockCacheList(ctxOther, cfg.SupportedLanguages)
			So(err, ShouldBeNil)

			DatasetPage(mockZebedeeClient, mockRend, mockFilesAPIClient, mockCacheList)(w, req)

			Convey("Then the status should be OK", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})

			Convey("And the response body should be the body returned from Zebedee", func() {
				actualBody, _ := ioutil.ReadAll(w.Body)
				So(actualBody, ShouldResemble, expectedBody)
			})
		})
	})
}

func TestDatasetTemplateRendering(t *testing.T) {
	cfg, _ := config.Get()
	renderClient := render.NewWithDefaultClient(assets.Asset, assets.AssetNames, cfg.PatternLibraryAssetsPath, "https://ons.gov.uk")

	extension := "csv"
	filename := "test" + "." + extension
	uri := "/path/to/" + filename
	versionFilename := "version_test.csv"

	Convey("Given a file stored in Zebedee", t, func() {
		expectedFormat := "/file?uri=%s"

		expectedDownloadUrl := fmt.Sprintf(expectedFormat, uri)
		expectedVersionDownloadUrl := fmt.Sprintf(expectedFormat, versionFilename)
		actualPageModel := mockZebedeePageModel(extension, uri, filename, versionFilename, expectedDownloadUrl, expectedVersionDownloadUrl)

		Convey("When render client page is built", func() {
			w, _ := generateRecorderAndRequest()
			renderClient.BuildPage(w, actualPageModel, "dataset")
			doc, _ := goquery.NewDocumentFromReader(w.Body)

			Convey("Then the download href should contain a /file path", func() {
				var actualHref string
				selector := fmt.Sprintf(`a[data-gtm-download-file="%s"]`, filename)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					actualHref, _ = s.Attr("href")
				})

				So(actualHref, ShouldEqual, expectedDownloadUrl)
			})

			Convey("Then the version download href should contain a /file path", func() {
				var actualHref string
				selector := fmt.Sprintf(`a[data-gtm-download-file="%s"]`, versionFilename)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					actualHref, _ = s.Attr("href")
				})

				So(actualHref, ShouldEqual, expectedVersionDownloadUrl)
			})
		})
	})

	Convey("Given a file stored in Files API", t, func() {
		expectedDownloadUrl := fmt.Sprintf("/%s%s", staticFilesDownloadEndpoint, uri)
		expectedVersionDownloadUrl := fmt.Sprintf("/%s/%s", staticFilesDownloadEndpoint, versionFilename)
		actualPageModel := mockZebedeePageModel(extension, uri, filename, versionFilename, expectedDownloadUrl, expectedVersionDownloadUrl)

		Convey("When render client page is built", func() {
			w, _ := generateRecorderAndRequest()
			renderClient.BuildPage(w, actualPageModel, "dataset")
			doc, _ := goquery.NewDocumentFromReader(w.Body)

			Convey("Then the download href should contain a /downloads-new path", func() {
				var actualHref string
				selector := fmt.Sprintf(`a[data-gtm-download-file="%s"]`, filename)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					actualHref, _ = s.Attr("href")
				})

				So(actualHref, ShouldEqual, expectedDownloadUrl)
			})

			Convey("Then the version download href should contain a /downloads-new path", func() {
				var actualHref string
				selector := fmt.Sprintf(`a[data-gtm-download-file="%s"]`, versionFilename)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					actualHref, _ = s.Attr("href")
				})

				So(actualHref, ShouldEqual, expectedVersionDownloadUrl)
			})
		})
	})
}

func mockZebedeePageModel(extension, uri, downloadFilename, versionFilename, downloadUrl, versionUrl string) mapper.DatasetPage {
	size := "100"
	basePath := "/some/path/2022"
	latestVersionUri := basePath + "/previous/v3"
	expectedDownload := dsp.Download{
		Extension:   extension,
		Size:        size,
		URI:         uri,
		File:        downloadFilename,
		DownloadURL: downloadUrl,
	}

	return mapper.DatasetPage{
		Page: coreModel.Page{
			SiteDomain: "https://foo.bar.com",
			Language:   "en",
		},
		DatasetPage: dsp.DatasetPage{
			Downloads: []dsp.Download{expectedDownload},
			Versions: []dsp.Version{
				{
					URI:       latestVersionUri,
					Downloads: []dsp.Download{{URI: versionFilename, DownloadURL: versionUrl, Extension: extension}},
				},
			},
		},
	}
}

func generateRecorderAndRequest() (*httptest.ResponseRecorder, *http.Request) {
	req, _ := http.NewRequest(http.MethodGet, editionPageURI, nil)
	req.Header.Add("X-Florence-Token", userAuthTokenDatasets)
	req.Header.Add("Collection-Id", collectionIDDatasets)
	localeCookie := &http.Cookie{
		Name:  "lang",
		Value: localeDatasets,
	}
	req.AddCookie(localeCookie)
	return httptest.NewRecorder(), req
}
