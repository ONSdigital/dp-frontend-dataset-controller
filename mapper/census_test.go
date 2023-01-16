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

// func TestCreateCensusDatasetLandingPage(t *testing.T) {
// helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
// req := httptest.NewRequest("", "/", nil)
// pageModel := coreModel.Page{}
// contacts := getTestContacts()
// relatedContent := getTestRelatedContent()
// datasetModel := getTestDatasetDetails(contacts, relatedContent)
// datasetOptions := getTestOptionsList()
// serviceMessage := getTestServiceMessage()
// emergencyBanner := getTestEmergencyBanner()
// datasetOptions := getTestOptionsList()
// 	versionOneDetails := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
// 	versionTwoDetails := getTestVersionDetails(2, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}),
// 		&[]dataset.Alert{
// 			{
// 				Date:        "",
// 				Description: "This is a correction",
// 				Type:        "correction",
// 			},
// 		})
// 	versionThreeDetails := getTestVersionDetails(4, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), &[]dataset.Alert{})

// 	filterDims := []sharedModel.FilterDimension{
// 		getTestFilterDimension("Geography", true, []string{"An option", "and another"}),
// 	}

// }

func TestCleanDimensionsLabel(t *testing.T) {

	Convey("Removes categories count from label - case insensitive", t, func() {
		So(cleanDimensionLabel("Example (100 categories)"), ShouldEqual, "Example")
		So(cleanDimensionLabel("Example (7 Categories)"), ShouldEqual, "Example")
		So(cleanDimensionLabel("Example (1 category)"), ShouldEqual, "Example")
		So(cleanDimensionLabel("Example (1 Category)"), ShouldEqual, "Example")
		So(cleanDimensionLabel(""), ShouldEqual, "")
		So(cleanDimensionLabel("Example 1 category"), ShouldEqual, "Example 1 category")
		So(cleanDimensionLabel("Example (something in brackets) (1 Category)"), ShouldEqual, "Example (something in brackets)")
	})

	Convey("Given simple page data", t, func() {
		helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
		req := httptest.NewRequest("", "/", nil)
		pageModel := coreModel.Page{}
		contact := getTestContacts()
		relatedContent := getTestRelatedContent()
		datasetModel := getTestDatasetDetails(contact, relatedContent)
		datasetOptions := getTestOptionsList()
		serviceMessage := getTestServiceMessage()
		emergencyBanner := getTestEmergencyBanner()

		Convey("and dimension labels that include category counts", func() {
			dimensions := []dataset.VersionDimension{
				{
					Description:          "A description on one line",
					Name:                 "dim_1",
					ID:                   "dim_1",
					Label:                "Label 1 (1 Category)",
					IsAreaType:           helpers.ToBoolPtr(true),
					QualityStatementText: "This is a quality notice statement",
					QualityStatementURL:  "#",
				},
				{
					Description:          "A description on one line \n Then a line break",
					Name:                 "dim_2",
					Label:                "Label 2 (100 categories)",
					ID:                   "dim_2",
					QualityStatementText: "This is another quality notice statement",
					QualityStatementURL:  "#",
				},
			}
			version := getTestVersionDetails(1, dimensions, getTestDownloads([]string{"xlsx"}), nil)

			Convey("when we build a dataset landing page", func() {
				page := CreateCensusLandingPage(true, context.Background(), req, pageModel, datasetModel, version, datasetOptions, "", false, []dataset.Version{version}, 1, "/a/version/1", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)

				Convey("then labels are formatted without counts", func() {
					So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Label 1")
					So(page.Collapsible.CollapsibleItems[3].Subheading, ShouldEqual, "Label 2")
					So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Label 1")
				})
			})
		})

		Convey("and filter dimension labels that include category counts", t, func() {
			filterDimensions := []sharedModel.FilterDimension{
				{
					ModelDimension: filter.ModelDimension{
						Label:      "Label 1 (100 categories)",
						Options:    []string{"An option", "and another"},
						IsAreaType: helpers.ToBoolPtr(true),
						Name:       "Geography",
					},
					OptionsCount: 2,
				},
			}
			Convey("when we build a dataset landing page", func() {
				page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, getTestVersionOneDetails(), datasetOptions, "", false, []dataset.Version{getTestVersionOneDetails()}, 1, "/a/version/1", "", []string{}, 50, false, true, true, getTestFilterDownloads([]string{"xlsx"}), filterDimensions, serviceMessage, emergencyBanner)

				Convey("then labels are formatted without counts", func() {
					So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, "Label 1")
				})
			})
		})
	})
}

func getTestRelatedContent() []dataset.GeneralDetails {
	return []dataset.GeneralDetails{
		{
			Title:       "Test related content 1",
			HRef:        "testrc1.example.com",
			Description: "Description of test related content 1",
		},
		{
			Title:       "Test related content 2",
			HRef:        "testrc2.example.com",
			Description: "Description of test related content 2",
		},
	}
}

func getTestContacts() []dataset.Contact {
	return []dataset.Contact{
		{
			Telephone: "01232 123 123",
			Email:     "hello@testing.com",
		},
	}
}

func getTestDatasetDetails(contacts []dataset.Contact, relatedContent []dataset.GeneralDetails) dataset.DatasetDetails {
	return dataset.DatasetDetails{
		Contacts:          &contacts,
		ID:                "12345",
		Description:       "An interesting test description \n with a line break",
		Title:             "Test title",
		Type:              "cantabular_flexible_table",
		NationalStatistic: true,
		Survey:            "census",
		RelatedContent:    &relatedContent,
	}
}

func getTestVersionDetails(versionNo int, dimensions []dataset.VersionDimension, downloads map[string]dataset.Download, alerts *[]dataset.Alert) dataset.Version {
	return dataset.Version{
		ReleaseDate: fmt.Sprintf("01-0%d-2021", versionNo),
		Downloads:   downloads,
		Edition:     "2021",
		Version:     versionNo,
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cantabular-1",
				ID:  "cantabular-1",
			},
		},
		Dimensions: dimensions,
		Alerts:     alerts,
	}
}

func getTestDimension(dimensionID string, isAreaType bool) dataset.VersionDimension {
	return dataset.VersionDimension{
		Description: fmt.Sprintf("A description for Dimension %s", dimensionID),
		Name:        fmt.Sprintf("Dimension %s", dimensionID),
		ID:          fmt.Sprintf("dim_%s", dimensionID),
		Label:       fmt.Sprintf("Label %s", dimensionID),
		IsAreaType:  helpers.ToBoolPtr(isAreaType),
	}
}

func getTestFilterDimension(name string, isAreaType bool, options []string) sharedModel.FilterDimension {
	return sharedModel.FilterDimension{
		ModelDimension: filter.ModelDimension{
			Label:      fmt.Sprintf("Label %s", name),
			Options:    options,
			IsAreaType: helpers.ToBoolPtr(isAreaType),
			Name:       name,
		},
		OptionsCount: len(options),
	}
}

func getTestDownloads(formats []string) map[string]dataset.Download {
	downloads := make(map[string]dataset.Download)
	for _, format := range formats {
		downloads[format] = dataset.Download{
			Size: "438290",
			URL:  "https://mydomain.com/my-request",
		}
	}
	return downloads
}

func getTestFilterDownloads(formats []string) map[string]filter.Download {
	downloads := make(map[string]filter.Download)
	for _, format := range formats {
		downloads[format] = filter.Download{
			Size: "438290",
			URL:  "https://mydomain.com/my-request",
		}
	}
	return downloads
}

func getTestDefaultDimensions() []dataset.VersionDimension {
	dim1 := getTestDimension("1", true)
	dim1.QualityStatementText = "This is a quality notice statement"
	dim1.QualityStatementURL = "#"

	dim2 := getTestDimension("2", false)
	dim2.QualityStatementText = "This is a another quality notice statement"
	dim2.QualityStatementURL = "#"

	dim3 := getTestDimension("3", false)
	dim3.Description = ""
	dim3.Name = "Only a name - I shouldn't map"

	return []dataset.VersionDimension{dim1, dim2, dim3}
}

func getTestOptions(dimensionID string, count int) dataset.Options {
	items := make([]dataset.Option, count)
	for i := 1; i <= 10; i++ {
		items = append(items, dataset.Option{
			DimensionID: dimensionID,
			Option:      fmt.Sprintf("option %d", i),
		})
	}

	return dataset.Options{
		Items:      items,
		TotalCount: count,
	}
}

func getTestOptionsList() []dataset.Options {
	return []dataset.Options{
		getTestOptions("dim_1", 2),
	}
}

func getTestVersionOneDetails() dataset.Version {
	return getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
}
