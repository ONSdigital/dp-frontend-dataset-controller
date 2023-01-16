package mapper

import (
	"context"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusFilterOutputsPage(t *testing.T) {
	Convey("Census dataset landing page maps correctly with filter output", t, func() {

		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, true, true, filterOutput, fDims, serviceMessage, emergencyBanner)
		So(page.Type, ShouldEqual, fmt.Sprintf("%s_filter_output", datasetModel.Type))

		So(page.SearchNoIndexEnabled, ShouldBeTrue)

		So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, fDims[0].Label)
		So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, fDims[0].Options)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[0].Name, ShouldEqual, fDims[0].Name)
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, fDims[0].Options)
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)

		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Coverage")

	})

	Convey("Downloads are properly mapped", t, func() {

		Convey("On filterOutput", func() {
			Convey("Where all four downloads exist", func() {
				page := CreateCensusDatasetLandingPage(
					true,
					context.Background(),
					req,
					pageModel,
					oneContactDetailDM,
					dataset.Version{},
					datasetOptions,
					"",
					false,
					[]dataset.Version{},
					1,
					"",
					"",
					[]string{},
					50,
					false,
					true,
					false,
					filterOutput,
					fDims,
					serviceMessage,
					emergencyBanner)

				Convey("HasDownloads set to true", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
				})

				Convey("Downloads are sorted by fixed extension order", func() {
					So(page.Version.Downloads[0].Extension, ShouldEqual, "xls")
					So(page.Version.Downloads[1].Extension, ShouldEqual, "csv")
					So(page.Version.Downloads[2].Extension, ShouldEqual, "txt")
					So(page.Version.Downloads[3].Extension, ShouldEqual, "csvw")
				})
			})
			Convey("Where excel download is missing", func() {
				page := CreateCensusDatasetLandingPage(
					true,
					context.Background(),
					req,
					pageModel,
					oneContactDetailDM,
					dataset.Version{},
					datasetOptions,
					"",
					false,
					[]dataset.Version{},
					1,
					"",
					"",
					[]string{},
					50,
					false,
					true,
					false,
					map[string]filter.Download{
						"csv": {
							Size: "12345",
							URL:  "https://mydomain.com/my-request",
						},
						"csvw": {
							Size: "12345",
							URL:  "https://mydomain.com/my-request",
						},
						"txt": {
							Size: "12345",
							URL:  "https://mydomain.com/my-request",
						},
					},
					fDims,
					serviceMessage,
					emergencyBanner)

				Convey("HasDownloads set to true", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
				})

				Convey("GetDataXLSXInfo set to true for custom datasets", func() {
					So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeTrue)
				})

				Convey("Downloads are sorted by fixed extension order", func() {
					So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
					So(page.Version.Downloads[1].Extension, ShouldEqual, "txt")
					So(page.Version.Downloads[2].Extension, ShouldEqual, "csvw")
				})
			})
			Convey("Where no downloads exist", func() {
				page := CreateCensusDatasetLandingPage(
					true,
					context.Background(),
					req,
					pageModel,
					oneContactDetailDM,
					dataset.Version{},
					datasetOptions,
					"",
					false,
					[]dataset.Version{},
					1,
					"",
					"",
					[]string{},
					50,
					false,
					true,
					false,
					map[string]filter.Download{},
					fDims,
					serviceMessage,
					emergencyBanner)

				Convey("HasDownloads set to false when downloads are zero", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
				})
			})

		})
	})

	Convey("given a dimension to truncate on filter output landing page", t, func() {
		fDims := []sharedModel.FilterDimension{
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 1",
					ID:    "dim_1",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
						"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
						"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
						"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
						"Label 21",
					},
					IsAreaType: helpers.ToBoolPtr(false),
				},
				OptionsCount: 21,
			},
			{
				ModelDimension: filter.ModelDimension{
					Label: "Dimension 2",
					ID:    "dim_2",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4",
						"Label 5", "Label 6", "Label 7", "Label 8",
						"Label 9", "Label 10", "Label 11", "Label 12",
					},
					IsAreaType: helpers.ToBoolPtr(false),
				},
				OptionsCount: 12,
			},
			{
				ModelDimension: filter.ModelDimension{
					Label:      "Dimension 3",
					ID:         "dim_3",
					IsAreaType: helpers.ToBoolPtr(true),
				},
				OptionsCount: 1,
			},
		}

		Convey("when valid parameters are provided", func() {
			p := CreateCensusDatasetLandingPage(
				true,
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{},
				50,
				false,
				true,
				false,
				filterOutput,
				fDims,
				serviceMessage,
				emergencyBanner)
			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 21)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
			})
		})

		Convey("when 'showAll' parameter provided", func() {
			p := CreateCensusDatasetLandingPage(
				true,
				context.Background(),
				req,
				pageModel,
				oneContactDetailDM,
				versionOneDetails,
				datasetOptions,
				"",
				false,
				[]dataset.Version{},
				1,
				"",
				"",
				[]string{"dim_1"},
				50,
				false,
				true,
				false,
				filterOutput,
				fDims,
				serviceMessage,
				emergencyBanner)
			Convey("then the dimension is no longer truncated", func() {
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 21)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeFalse)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/#dim_1")
				So(p.DatasetLandingPage.Dimensions[3].TotalItems, ShouldEqual, 12)
				So(p.DatasetLandingPage.Dimensions[3].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[3].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 5", "Label 6", "Label 7",
					"Label 10", "Label 11", "Label 12",
				})
				So(p.DatasetLandingPage.Dimensions[3].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[3].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})

	Convey("Analytics for Filter Outputs Pages are properly mapped", t, func() {
		Convey("given we have changed area_type only", func() {
			fDims := []sharedModel.FilterDimension{
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
				analytics := getFilterAnalytics(fDims, defaultCoverage)

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

		Convey("given we have changed area_type and have three coverage items at area_type level", func() {
			fDims := []sharedModel.FilterDimension{
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
				analytics := getFilterAnalytics(fDims, defaultCoverage)

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

		Convey("given we have more than four coverage items at area_type level", func() {
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
				Convey("but coverage is not set ", func() {
					_, ok := analytics["coverage"]
					So(ok, ShouldBeFalse)
				})
			})
		})

		Convey("given we are doing coverage using areas within a larger area", func() {
			fDims := []sharedModel.FilterDimension{
				{
					ModelDimension: filter.ModelDimension{
						ID:             "area_type_ID",
						IsAreaType:     helpers.ToBoolPtr(true),
						FilterByParent: "parent_area_type_ID",
						Options: []string{
							"Area1", "Area2", "Area3",
							``},
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

	})
}
