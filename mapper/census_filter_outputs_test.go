package mapper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/contact"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusFilterOutputsPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	// Setting `Type` here as this done by `UpdateBasePage` mapper when called in the handler
	pageModel := coreModel.Page{
		Type: "cantabular_flexible_table",
	}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	dimDesc := population.GetDimensionsResponse{
		Dimensions: []population.Dimension{
			{
				ID:          "geography",
				Label:       "Label geography",
				Description: "a geography description",
			},
			{
				ID:          "non-geog",
				Label:       "Label non-geog",
				Description: "a non-geography description",
			},
		},
	}

	Convey("given data for a census landing page with version 1", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{
			getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2),
			getTestFilterDimension("non-geog", false, []string{"option a", "option b"}, 2)}
		filterOutputs := filter.Model{
			Downloads:      getTestFilterDownloads([]string{"xlsx"}),
			PopulationType: "UR",
		}
		population := population.GetPopulationTypeResponse{
			PopulationType: population.PopulationType{
				Name:        "UR",
				Label:       "Usual residents",
				Description: "A description about usual residents",
			},
		}

		Convey("when we build a filter outputs page", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, dimDesc, cantabular.GetBlockedAreaCountResult{}, population)

			Convey("then the type should have _filter_output appended", func() {
				So(page.Type, ShouldEqual, fmt.Sprintf("%s_filter_output", datasetModel.Type))
			})
			Convey("and search should be disabled", func() {
				So(page.SearchNoIndexEnabled, ShouldBeTrue)
			})

			Convey("and dimensions map correctly", func() {
				So(page.DatasetLandingPage.Dimensions[0].IsPopulationType, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, filterDims[0].Label)
				So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, filterDims[0].Options)
				So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[1].Name, ShouldEqual, filterDims[0].Name)
				So(page.DatasetLandingPage.Dimensions[2].IsCoverage, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, filterDims[0].Options)
				So(page.DatasetLandingPage.Dimensions[2].ShowChange, ShouldBeTrue)
			})

			Convey("and collapsible items are ordered correctly", func() {
				So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
				So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, dimDesc.Dimensions[0].Label)
				So(page.Collapsible.CollapsibleItems[1].Content, ShouldResemble, []string{dimDesc.Dimensions[0].Description})
				So(page.Collapsible.CollapsibleItems[2].Subheading, ShouldEqual, "Coverage")
				So(page.Collapsible.CollapsibleItems[3].Subheading, ShouldEqual, dimDesc.Dimensions[1].Label)
				So(page.Collapsible.CollapsibleItems[3].Content, ShouldResemble, []string{dimDesc.Dimensions[1].Description})
			})
		})
	})

	Convey("test IsChangeVisible parameter", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{
			getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2),
			getTestFilterDimension("one-cat", false, []string{"option 1", "option 2"}, 1),
			getTestFilterDimension("two-cats", false, []string{"option 1", "option 2"}, 2),
		}
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx"}),
		}

		Convey("when isMultivariate is false", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, false, dimDesc, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then ShowChange is false for all", func() {
				So(page.DatasetLandingPage.Dimensions[3].ShowChange, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[4].ShowChange, ShouldBeFalse)
			})
		})

		Convey("when isMultivariate is true", func() {
			multivariateModel := getTestDatasetDetails(contacts, relatedContent)
			multivariateModel.Type = "cantabular_multivariate_table"
			page := CreateCensusFilterOutputsPage(req, pageModel, multivariateModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, dimDesc, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})
			Convey("then IsChangeCategories is false if categorisation is only one available", func() {
				So(page.DatasetLandingPage.Dimensions[3].ShowChange, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[4].ShowChange, ShouldBeTrue)
			})
		})
	})
}

func TestSDCOnFilterOutputsPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	datasetModel.Type = "multivariate" //nolint:goconst // not necessary for tests
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("given a request for a filter outputs census landing page", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2), getTestFilterDimension("non-geog", false, []string{"option a", "option b"}, 2)}
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx"}),
		}
		sdc := cantabular.GetBlockedAreaCountResult{
			Passed:  0,
			Blocked: 10,
			Total:   10,
		}

		Convey("when areas are blocked", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, sdc, population.GetPopulationTypeResponse{})

			Convey("then the sdc panel is displayed", func() {
				So(page.DatasetLandingPage.HasSDC, ShouldBeTrue)
			})
			Convey("then the panel type is 'pending'", func() {
				So(page.DatasetLandingPage.SDC[0].Type, ShouldEqual, census.Pending)
			})
		})

		Convey("when all areas are passing", func() {
			sdc = cantabular.GetBlockedAreaCountResult{
				Passed:  10,
				Blocked: 0,
				Total:   10,
			}
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, sdc, population.GetPopulationTypeResponse{})

			Convey("then the sdc panel is displayed", func() {
				So(page.DatasetLandingPage.HasSDC, ShouldBeTrue)
			})
			Convey("then the panel type is 'pending'", func() {
				So(page.DatasetLandingPage.SDC[0].Type, ShouldEqual, census.Success)
			})
		})
	})
}

func TestCustomHeadingOnFilterOutputs(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	// Setting `Metadata.Title` here as this done by `UpdateBasePage` mapper when called in the handler
	pageModel := coreModel.Page{
		Metadata: coreModel.Metadata{
			Title: "Test title",
		},
	}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	datasetModel.Type = "multivariate" //nolint:goconst //this string doesn't need to be a constant for tests
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("given a request for a filter outputs census landing page", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2), getTestFilterDimension("first", false, []string{}, 2), getTestFilterDimension("second", false, []string{}, 2)}
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx"}),
		}

		Convey("when the filter is a customised multivariate", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then the title is customised", func() {
				So(page.Metadata.Title, ShouldEqual, "Label first and label second")
			})
		})

		Convey("when the filter is multivariate and has not been customised", func() {
			filterDims = []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{}, 0), getTestFilterDimension("2", false, []string{}, 0), getTestFilterDimension("3", false, []string{}, 0)}
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then isCustom bool is set", func() {
				So(page.DatasetLandingPage.IsCustom, ShouldBeFalse)
			})
			Convey("then the title is not customised", func() {
				So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
			})
		})

		Convey("when the filter is a custom", func() {
			filterOutputs.Custom = helpers.ToBoolPtr(true)
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then isCustom bool is set", func() {
				So(page.DatasetLandingPage.IsCustom, ShouldBeTrue)
			})
			Convey("then the title is customised", func() {
				So(page.Metadata.Title, ShouldEqual, "Label first and label second")
			})
		})
	})
}

func TestMetadataOverridesOnCustomFilterOutputs(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	datasetModel.Type = "multivariate"
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("given a request for a filter outputs census landing page", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2), getTestFilterDimension("first", false, []string{}, 2), getTestFilterDimension("second", false, []string{}, 2)}
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx"}),
			Custom:    helpers.ToBoolPtr(true),
		}

		Convey("when the filter is custom", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then the title is customised", func() {
				So(page.Metadata.Title, ShouldEqual, "Label first and label second")
			})
			Convey("then the summary is set", func() {
				So(page.DatasetLandingPage.Description, ShouldResemble, []string{"This is a custom dataset"})
			})
			Convey("then the dataset id is blank", func() {
				So(page.DatasetId, ShouldBeBlank)
			})
			Convey("then the release date is blank", func() {
				So(page.ReleaseDate, ShouldBeBlank)
			})
			Convey("then the contact details are set", func() {
				So(page.ContactDetails, ShouldResemble, contact.Details{
					Email:     "census.customerservices@ons.gov.uk",
					Telephone: "+44 1329 444972",
				})
				So(page.HasContactDetails, ShouldBeTrue)
			})
			Convey("then the census branding is displayed", func() {
				So(page.ShowCensusBranding, ShouldBeTrue)
			})
			Convey("then the national statistic is displayed", func() {
				So(page.IsNationalStatistic, ShouldBeTrue)
			})
		})
	})
}

func TestCreateCensusFilterOutputsDownloads(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
	filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2)}

	Convey("given filter outputs where all four file types exist", t, func() {
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx", "txt", "csv", "csvw"}),
		}

		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then HasDownloads set to true when downloads are greater than three or more", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})

			Convey("and ShowXLSXInfo is set to true", func() {
				So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeTrue)
			})

			Convey("and downloads are sorted by fixed extension order", func() {
				So(page.Version.Downloads[0].Extension, ShouldEqual, "xlsx")
				So(page.Version.Downloads[1].Extension, ShouldEqual, "csv")
				So(page.Version.Downloads[2].Extension, ShouldEqual, "txt")
				So(page.Version.Downloads[3].Extension, ShouldEqual, "csvw")
			})
		})
	})

	Convey("given filter outputs where excel download is missing", t, func() {
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"txt", "csv", "csvw"}),
		}

		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then HasDownloads set to true when downloads are greater than three or more", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})

			Convey("and ShowXLSXInfo is set to false", func() {
				So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeTrue)
			})

			Convey("and downloads are sorted by fixed extension order", func() {
				So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
				So(page.Version.Downloads[1].Extension, ShouldEqual, "txt")
				So(page.Version.Downloads[2].Extension, ShouldEqual, "csvw")
			})
		})
	})

	Convey("given no downloads exist", t, func() {
		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filter.Model{}, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then HasDownloads set to false", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
	})
}

func TestCreateCensusFilterOutputsPagination(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
	filterOutputs := filter.Model{
		Downloads: getTestFilterDownloads([]string{"xlsx", "txt", "csv", "csvw"}),
	}

	Convey("given a dimension to truncate on filter output landing page", t, func() {
		filterDimensions := []sharedModel.FilterDimension{
			buildTestFilterDimension("dim_1", false, 21),
			buildTestFilterDimension("dim_2", false, 12),
			buildTestFilterDimension("dim_3", true, 1),
		}

		Convey("when valid parameters are provided", func() {
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDimensions, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(page.DatasetLandingPage.Dimensions[3].TotalItems, ShouldEqual, 21)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[3].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[3].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			parameters := []string{"dim_1"}
			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", parameters, false, true, filterOutputs, filterDimensions, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then the dimension is no longer truncated", func() {
				So(page.DatasetLandingPage.Dimensions[3].TotalItems, ShouldEqual, 21)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[3].IsTruncated, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[3].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[3].TruncateLink, ShouldEqual, "/#dim_1")
				So(page.DatasetLandingPage.Dimensions[4].TotalItems, ShouldEqual, 12)
				So(page.DatasetLandingPage.Dimensions[4].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[4].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 5", "Label 6", "Label 7",
					"Label 10", "Label 11", "Label 12",
				})
				So(page.DatasetLandingPage.Dimensions[4].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[4].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})
}

func TestCreateCensusFilterOutputsQualityNotices(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	datasetModel.Type = "multivariate"
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("given a request for a filter outputs census landing page", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"}, 2), getTestFilterDimension("first", false, []string{}, 2), getTestFilterDimension("second", false, []string{}, 2)}
		filterOutputs := filter.Model{
			Downloads: getTestFilterDownloads([]string{"xlsx"}),
		}

		Convey("when there is a quality notice on the dimension", func() {
			filterDims[0].QualityStatementText = "This is a quality notice statement"
			filterDims[0].QualitySummaryURL = "https://quality-notice-1.com"
			filterDims[1].QualityStatementText = "This is another quality notice statement"
			filterDims[1].QualitySummaryURL = "https://quality-notice-2.com"

			page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, version, false, []dpDatasetApiModels.Version{version}, 1, "/a/version/1", "", []string{}, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

			Convey("then the 'quality notice' panel is displayed", func() {
				mockPanel := []census.Panel{
					{
						Body:       []string{"<p>This is a quality notice statement</p>Read more about this"},
						CSSClasses: []string{"ons-u-mt-no"},
					},
					{
						Body:       []string{"<p>This is another quality notice statement</p>Read more about this"},
						CSSClasses: []string{"ons-u-mt-no", "ons-u-mb-l"},
					},
				}
				So(page.DatasetLandingPage.QualityStatements, ShouldHaveLength, 2)
				So(page.DatasetLandingPage.QualityStatements, ShouldResemble, mockPanel)
			})
		})
	})
}

func TestCreateCensusFilterOutputsAnalytics(t *testing.T) {
	Convey("given we have changed area_type only", t, func() {
		filterDimensions := []sharedModel.FilterDimension{
			{
				ModelDimension: filter.ModelDimension{
					ID:         "area_type_ID",
					IsAreaType: helpers.ToBoolPtr(true),
				},
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 1",
					ID:    "dimension_ID_1",
				},
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 2",
					ID:    "dimension_ID_2",
				},
			},
		}
		defaultCoverage := true

		Convey("when we generate analytics data", func() {
			analytics := getFilterAnalytics(filterDimensions, defaultCoverage)

			Convey("then and areatype should be set to the area dimension ID", func() {
				So(analytics["areaType"], ShouldEqual, "area_type_ID")
			})
			Convey("and coverage count is zero", func() {
				So(analytics["coverageCount"], ShouldEqual, "0")
			})
			Convey("and coverage is not set", func() {
				_, ok := analytics["coverage"]
				So(ok, ShouldBeFalse)
			})
			Convey("and coverageAreaType is not set", func() {
				_, ok := analytics["coverageAreaType"]
				So(ok, ShouldBeFalse)
			})
			Convey("and dimensions should exclude IsAreaType", func() {
				So(analytics["dimensions"], ShouldEqual, "dimension_ID_1,dimension_ID_2")
			})
		})
	})

	Convey("given we have changed area_type and have three coverage items at area_type level", t, func() {
		filterDimensions := []sharedModel.FilterDimension{
			{
				ModelDimension: filter.ModelDimension{
					ID:         "area_type_ID",
					IsAreaType: helpers.ToBoolPtr(true),
					Options: []string{
						"Area1", "Area2", "Area3",
					},
				},
				OptionsCount: 3,
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 1",
					ID:    "dimension_ID_1",
				},
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 2",
					ID:    "dimension_ID_2",
				},
			},
		}
		defaultCoverage := false

		Convey("when we generate analytics data", func() {
			analytics := getFilterAnalytics(filterDimensions, defaultCoverage)

			Convey("then areatype should be set to the area dimension ID", func() {
				So(analytics["areaType"], ShouldEqual, "area_type_ID")
			})
			Convey("and coverage count is three", func() {
				So(analytics["coverageCount"], ShouldEqual, "3")
			})
			Convey("and coverage is set to the comma joined list of areas", func() {
				So(analytics["coverage"], ShouldEqual, "Area1,Area2,Area3")
			})
			Convey("and coverageAreaType is set to the area dimension ID", func() {
				So(analytics["coverageAreaType"], ShouldEqual, "area_type_ID")
			})
			Convey("and dimensions should exclude IsAreaType", func() {
				So(analytics["dimensions"], ShouldEqual, "dimension_ID_1,dimension_ID_2")
			})
		})
	})

	Convey("given we have more than four coverage items at area_type level", t, func() {
		fDims := []sharedModel.FilterDimension{
			{
				ModelDimension: filter.ModelDimension{
					ID:         "area_type_ID",
					IsAreaType: helpers.ToBoolPtr(true),
					Options: []string{
						"Area1", "Area2", "Area3", "Area4", "Area5",
					},
				},
				OptionsCount: 3,
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 1",
					ID:    "dimension_ID_1",
				},
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 2",
					ID:    "dimension_ID_2",
				},
			},
		}
		defaultCoverage := false

		Convey("when we generate analytics data", func() {
			analytics := getFilterAnalytics(fDims, defaultCoverage)

			Convey("then coverage count is set", func() {
				So(analytics["coverageCount"], ShouldEqual, "5")
			})
			Convey("and coverageAreaType is set", func() {
				So(analytics["coverageAreaType"], ShouldEqual, "area_type_ID")
			})
			Convey("and coverage is not set ", func() {
				_, ok := analytics["coverage"]
				So(ok, ShouldBeFalse)
			})
		})
	})

	Convey("given we are doing coverage using areas within a larger area", t, func() {
		fDims := []sharedModel.FilterDimension{
			{
				ModelDimension: filter.ModelDimension{
					ID:             "area_type_ID",
					IsAreaType:     helpers.ToBoolPtr(true),
					FilterByParent: "parent_area_type_ID",
					Options:        []string{"Area1", "Area2", "Area3"},
				},
				OptionsCount: 3,
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 1",
					ID:    "dimension_ID_1",
				},
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 2",
					ID:    "dimension_ID_2",
				},
			},
		}
		defaultCoverage := false

		Convey("when we generate analytics data", func() {
			analytics := getFilterAnalytics(fDims, defaultCoverage)

			Convey("then coverageAreaType is set to the larger area", func() {
				So(analytics["coverageAreaType"], ShouldEqual, "parent_area_type_ID")
			})
		})
	})
}

func TestMapBlockedAreasPanel(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	tc := []struct {
		sdc       cantabular.GetBlockedAreaCountResult
		panelType census.PanelType
		expected  []census.Panel
	}{
		{
			sdc: cantabular.GetBlockedAreaCountResult{
				Passed:  10,
				Blocked: 15,
				Total:   25,
			},
			panelType: census.Pending,
			expected: []census.Panel{
				{
					Type:        census.Pending,
					DisplayIcon: false,
					CSSClasses:  []string{"ons-u-mt-xl", "ons-u-mb-s"},
					Body:        []string{"10 out of 25 areas available", "Protecting personal data will prevent 15 areas from being published"},
					Language:    "en",
				},
			},
		},
		{
			sdc: cantabular.GetBlockedAreaCountResult{
				Passed:  10,
				Blocked: 0,
				Total:   10,
			},
			panelType: census.Success,
			expected: []census.Panel{
				{
					Type:        census.Success,
					DisplayIcon: false,
					CSSClasses:  []string{"ons-u-mt-xl", "ons-u-mb-s"},
					Body:        []string{"All 10 areas available"},
					Language:    "en",
				},
			},
		},
		{
			sdc:       cantabular.GetBlockedAreaCountResult{},
			panelType: 0,
			expected:  []census.Panel(nil),
		},
	}

	Convey("Given a list", t, func() {
		Convey("When the mapBlockedAreasPanel function is called", func() {
			for i, test := range tc {
				Convey(fmt.Sprintf("Then the given parameters in test index %d returns the expected result", i), func() {
					So(mapBlockedAreasPanel(&test.sdc, test.panelType, "en"), ShouldResemble, test.expected)
				})
			}
		})
	})
}

func TestMapImproveResultsCollapsible(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	mockDims := []sharedModel.Dimension{
		{
			Name:       "Area type",
			IsAreaType: true,
			IsCoverage: false,
		},
		{
			Name:       "Coverage",
			IsAreaType: false,
			IsCoverage: true,
		},
		{
			Name:       "Another dimension",
			IsAreaType: false,
			IsCoverage: false,
		},
	}
	mockCollapsible := coreModel.Collapsible{
		Title: coreModel.Localisation{
			LocaleKey: "ImproveResultsTitle",
			Plural:    4,
		},
		CollapsibleItems: []coreModel.CollapsibleItem{
			{
				Subheading: "Try the following",
				Content:    []string(nil),
				SafeHTML: coreModel.Localisation{
					Text: "A list of suggestions",
				},
			},
		},
	}

	Convey("Given a list", t, func() {
		Convey("When the mapImproveResultsCollapsible function is called", func() {
			Convey("Then the given dimensions returns the expected collapsible", func() {
				So(mapImproveResultsCollapsible(mockDims, "en"), ShouldResemble, mockCollapsible)
			})
		})
	})
}

func TestBuildDimsList(t *testing.T) {
	tc := []struct {
		given    []string
		useAnd   bool
		expected string
	}{
		{
			given:    []string{},
			expected: "",
		},
		{
			given: []string{
				"a human name",
			},
			expected: "a human name",
		},
		{
			given: []string{
				"a human name",
				"another human name",
			},
			expected: "a human name or another human name",
		},
		{
			given: []string{
				"a human name",
				"another human name",
			},
			useAnd:   true,
			expected: "a human name and another human name",
		},
		{
			given: []string{
				"a human name",
				"another human name",
				"this human name",
			},
			expected: "a human name, another human name or this human name",
		},
		{
			given: []string{
				"a human name",
				"another human name",
				"this human name",
				"human name",
			},
			expected: "a human name, another human name, this human name or human name",
		},
		{
			given: []string{
				"a human name",
				"another human name",
				"this human name",
				"human name",
			},
			useAnd:   true,
			expected: "a human name, another human name, this human name and human name",
		},
	}

	Convey("Given a list", t, func() {
		Convey("When the buildDimsList function is called", func() {
			for i, test := range tc {
				Convey(fmt.Sprintf("Then the given list (test index %d) returns %s", i, test.expected), func() {
					So(buildConjoinedList(test.given, test.useAnd), ShouldEqual, test.expected)
				})
			}
		})
	})
}
