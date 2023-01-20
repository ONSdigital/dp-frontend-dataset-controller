package mapper

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusFilterOutputsPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("given data for a census landing page with version 1", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
		filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"})}
		filterOutputs := getTestFilterDownloads([]string{"xlsx"})

		Convey("when we build a filter outputs page", func() {
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true)

			Convey("then the type should have _filter_output appended", func() {
				So(page.Type, ShouldEqual, fmt.Sprintf("%s_filter_output", datasetModel.Type))
			})
			Convey("and search should be disabled", func() {
				So(page.SearchNoIndexEnabled, ShouldBeTrue)
			})

			Convey("and dimensions map correctly", func() {
				So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, filterDims[0].Label)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, filterDims[0].Options)
				So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[0].Name, ShouldEqual, filterDims[0].Name)
				So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, filterDims[0].Options)
				So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)
			})

			Convey("and collapsible items are ordered correctly", func() {
				So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
				So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Coverage")
			})
		})
	})
}

func TestCreateCensusFilterOutputsDownloads(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
	filterDims := []sharedModel.FilterDimension{getTestFilterDimension("geography", true, []string{"option 1", "option 2"})}

	Convey("given filter outputs where all four file types exist", t, func() {
		filterOutputs := getTestFilterDownloads([]string{"xlsx", "txt", "csv", "csvw"})

		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true)

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
		filterOutputs := getTestFilterDownloads([]string{"txt", "csv", "csvw"})

		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true)

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
		filterOutputs := map[string]filter.Download{}

		Convey("when we build a census landing page", func() {
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, true, filterOutputs, filterDims, serviceMessage, emergencyBanner, true)

			Convey("then HasDownloads set to false", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
	})

}

func TestCreateCensusFilterOutputsPagination(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
	filterOutputs := getTestFilterDownloads([]string{"xlsx", "txt", "csv", "csvw"})

	Convey("given a dimension to truncate on filter output landing page", t, func() {
		filterDimensions := []sharedModel.FilterDimension{
			buildTestFilterDimension("dim_1", false, 21),
			buildTestFilterDimension("dim_2", false, 12),
			buildTestFilterDimension("dim_3", true, 1),
		}

		Convey("when valid parameters are provided", func() {
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, true, filterOutputs, filterDimensions, serviceMessage, emergencyBanner, true)

			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(page.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 21)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			parameters := []string{"dim_1"}
			page := CreateCensusFilterOutputsPage(context.Background(), req, pageModel, datasetModel, version, "", false, []dataset.Version{version}, 1, "/a/version/1", "", parameters, 50, false, true, filterOutputs, filterDimensions, serviceMessage, emergencyBanner, true)

			Convey("then the dimension is no longer truncated", func() {
				So(page.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 21)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/#dim_1")
				So(page.DatasetLandingPage.Dimensions[3].TotalItems, ShouldEqual, 12)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[3].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 5", "Label 6", "Label 7",
					"Label 10", "Label 11", "Label 12",
				})
				So(page.DatasetLandingPage.Dimensions[3].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[3].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
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
