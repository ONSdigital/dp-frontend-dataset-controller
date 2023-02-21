package mapper

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusLandingPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	datasetOptions := getTestOptionsList()
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("Given a census landing page version 1", t, func() {
		version := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)

		Convey("When we build a census landing page", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			Convey("Then downloads map correctly", func() {
				So(page.Version.Downloads[0].Size, ShouldEqual, "438290")
				So(page.Version.Downloads[0].Extension, ShouldEqual, "xlsx")
				So(page.Version.Downloads[0].URI, ShouldEqual, "https://mydomain.com/my-request")
				So(page.Version.Downloads, ShouldHaveLength, 1)

			})

			Convey("And Dimensions map correctly", func() {
				So(page.DatasetLandingPage.Dimensions, ShouldHaveLength, 2) // coverage is inserted
				So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Coverage")
				So(page.DatasetLandingPage.Dimensions[1].Name, ShouldEqual, "coverage")
				So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeFalse)
			})

			Convey("And collapsible items are mapped from dimensions", func() {
				So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
				So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, version.Dimensions[0].Label)
				So(page.Collapsible.CollapsibleItems[1].Content[0], ShouldEqual, version.Dimensions[0].Description)
				So(page.Collapsible.CollapsibleItems[2].Subheading, ShouldEqual, "Coverage")
				So(page.Collapsible.CollapsibleItems[3].Subheading, ShouldEqual, version.Dimensions[1].Label)
				So(page.Collapsible.CollapsibleItems[3].Content, ShouldResemble, strings.Split(version.Dimensions[1].Description, "\n"))
				So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 4)
			})

			Convey("And the page should appear in search", func() {
				So(page.SearchNoIndexEnabled, ShouldBeFalse)
			})
		})
	})
}

func TestCreateCensusLandingPageQualityNotices(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("Given there are quality notices on dimensions", t, func() {
		dim1 := getTestDimension("1", true)
		dim1.QualityStatementText = "This is a quality notice statement"
		dim1.QualityStatementURL = "#"

		dim2 := getTestDimension("2", false)
		dim2.QualityStatementText = "This is another quality notice statement"
		dim2.QualityStatementURL = "#"

		dim3 := getTestDimension("3", false)
		dim3.Description = ""
		dim3.Name = "Only a name - I shouldn't map"

		dimensions := []dataset.VersionDimension{dim1, dim2, dim3}
		version := getTestVersionDetails(1, dimensions, getTestDownloads([]string{"xlsx"}), nil)
		datasetOptions := []dataset.Options{
			getTestOptions("Dimension 1", 1),
			getTestOptions("Dimension 2", 1),
		}

		Convey("When we build a census landing page", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			mockPanel := []datasetLandingPageCensus.Panel{
				{
					Body:       []string{"<p>This is a quality notice statement</p>Read more about this"},
					CssClasses: []string{"ons-u-mt-no"},
				},
				{
					Body:       []string{"<p>This is another quality notice statement</p>Read more about this"},
					CssClasses: []string{"ons-u-mt-no", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'quality notice' panel is displayed", func() {
				So(page.DatasetLandingPage.QualityStatements, ShouldHaveLength, 2)
				So(page.DatasetLandingPage.QualityStatements, ShouldResemble, mockPanel)
			})
		})
	})
}

func TestCreateCensusLandingPageDownloads(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	datasetOptions := getTestOptionsList()

	Convey("given download data where all four file types exist", t, func() {
		downloads := getTestDownloads([]string{"csv", "xls", "csvw", "txt"})
		version := getTestVersionDetails(1, getTestDefaultDimensions(), downloads, nil)

		Convey("when we build a census landing page", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			Convey("then HasDownloads set to true when downloads are greater than three or more", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})

			Convey("and ShowXLSXInfo is set to false", func() {
				So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeFalse)
			})

			Convey("and downloads are sorted by fixed extension order", func() {
				So(page.Version.Downloads[0].Extension, ShouldEqual, "xls")
				So(page.Version.Downloads[1].Extension, ShouldEqual, "csv")
				So(page.Version.Downloads[2].Extension, ShouldEqual, "txt")
				So(page.Version.Downloads[3].Extension, ShouldEqual, "csvw")
			})
		})
	})

	Convey("given download data where the excel file is missing", t, func() {
		downloads := getTestDownloads([]string{"csv", "csvw", "txt"})
		version := getTestVersionDetails(1, getTestDefaultDimensions(), downloads, nil)

		Convey("when we build a census landing page", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			Convey("then HasDownloads set to true when downloads are greater than three or more", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})

			Convey("and ShowXLSXInfo is set to false", func() {
				So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeFalse)
			})

			Convey("and downloads are sorted by fixed extension order", func() {
				So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
				So(page.Version.Downloads[1].Extension, ShouldEqual, "txt")
				So(page.Version.Downloads[2].Extension, ShouldEqual, "csvw")
			})
		})
	})

	Convey("given download data where the excel file is missing", t, func() {
		downloads := map[string]dataset.Download{}
		version := getTestVersionDetails(1, getTestDefaultDimensions(), downloads, nil)

		Convey("when we build a census landing page", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			Convey("then HasDownloads set to false", func() {
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
	})
}

func TestCreateCensusLandingPagePagination(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	version := getTestVersionOneDetails()

	Convey("given a dimension to truncate on census dataset landing page", t, func() {
		datasetOptions := []dataset.Options{
			getTestOptions("Dimension 1", 21),
			getTestOptions("Dimension 2", 20),
		}

		Convey("when valid parameters are provided", func() {
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, serviceMessage, emergencyBanner, true)

			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(page.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			parameters := []string{"dim_1"}
			page := CreateCensusLandingPage(context.Background(), req, pageModel, datasetModel, version, datasetOptions, map[string]int{}, "", false, []dataset.Version{version}, 1, "/a/version/1", "", parameters, 50, false, serviceMessage, emergencyBanner, true)

			Convey("then the dimension is no longer truncated", func() {
				So(page.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeFalse)
				So(page.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(page.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
				So(page.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, datasetOptions[1].TotalCount)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(page.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 18", "Label 19", "Label 20",
				})
				So(page.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(page.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})
}
func TestCreateCensusLandingAnalytics(t *testing.T) {
	Convey("given dimension data for a dataset landing page", t, func() {
		dimensions := []model.Dimension{
			{
				ID:         "area_ID",
				IsAreaType: true,
				IsCoverage: false,
			},
			{
				ID:         "coverage",
				IsAreaType: false,
				IsCoverage: true,
			},
			{
				ID:         "dimension_ID_1",
				IsAreaType: false,
				IsCoverage: false,
			},
			{
				ID:         "dimension_ID_2",
				IsAreaType: false,
				IsCoverage: false,
			},
		}

		Convey("when we generate analytics data", func() {
			analytics := getAnalytics(dimensions)

			Convey("then coverage count is zero", func() {
				So(analytics["coverageCount"], ShouldEqual, "0")
			})
			Convey("and areatype should be set to the area dimension ID", func() {
				So(analytics["areaType"], ShouldEqual, "area_ID")
			})
			Convey("and dimensions should exclude IsAreaType or IsCoverage dimensions", func() {
				So(analytics["dimensions"], ShouldEqual, "dimension_ID_1,dimension_ID_2")
			})
		})
	})
}
