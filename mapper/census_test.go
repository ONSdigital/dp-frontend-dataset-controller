package mapper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	dpDatasetApiModels "github.com/ONSdigital/dp-dataset-api/models"
	dpDatasetApiSdk "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

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
		req := httptest.NewRequest("", "/", http.NoBody)
		pageModel := coreModel.Page{}
		contact := getTestContacts()
		relatedContent := getTestRelatedContent()
		datasetModel := getTestDatasetDetails(contact, relatedContent)
		datasetOptions := []dpDatasetApiSdk.VersionDimensionOptionsList{
			getTestOptions("dim_1", 10),
			getTestOptions("dim_2", 10),
		}
		serviceMessage := getTestServiceMessage()
		emergencyBanner := getTestEmergencyBanner()

		Convey("and dimension labels that include category counts", func() {
			dimensions := []dpDatasetApiModels.Dimension{
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
				page := CreateCensusLandingPage(pageModel, datasetModel, version, datasetOptions, map[string]int{},
					[]dpDatasetApiModels.Version{version}, []string{}, true, population.GetPopulationTypeResponse{})

				Convey("then labels are formatted without counts", func() {
					So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Label 1")
					So(page.Collapsible.CollapsibleItems[3].Subheading, ShouldEqual, "Label 2")
					So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Label 1")
				})
			})
		})

		Convey("and filter dimension labels that include category counts", func() {
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
				page := CreateCensusFilterOutputsPage(req, pageModel, datasetModel, getTestVersionOneDetails(), false, []dpDatasetApiModels.Version{getTestVersionOneDetails()}, 1, "/a/version/1", "", []string{}, false, true, filter.Model{Downloads: getTestFilterDownloads([]string{"xlsx"})}, filterDimensions, serviceMessage, emergencyBanner, true, population.GetDimensionsResponse{}, cantabular.GetBlockedAreaCountResult{}, population.GetPopulationTypeResponse{})

				Convey("then labels are formatted without counts", func() {
					So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Label 1")
				})
			})
		})
	})
}

func getTestRelatedContent() []dpDatasetApiModels.GeneralDetails {
	return []dpDatasetApiModels.GeneralDetails{
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

func getTestContacts() []dpDatasetApiModels.ContactDetails {
	return []dpDatasetApiModels.ContactDetails{
		{
			Telephone: "01232 123 123",
			Email:     "hello@testing.com",
		},
	}
}

// Returns a representative populated `dp-dataset-api.models.Dataset` data struct
func getTestDatasetDetails(contacts []dpDatasetApiModels.ContactDetails, relatedContent []dpDatasetApiModels.GeneralDetails) dpDatasetApiModels.Dataset {
	var nationalStatistic = true

	return dpDatasetApiModels.Dataset{
		Contacts:          contacts,
		ID:                "cantabular-1",
		Description:       "An interesting test description \n with a line break",
		Title:             "Test title",
		Type:              "cantabular_flexible_table",
		NationalStatistic: &nationalStatistic,
		Survey:            "census",
		RelatedContent:    relatedContent,
		IsBasedOn: &dpDatasetApiModels.IsBasedOn{
			ID: "UR",
		},
	}
}

func getTestVersionDetails(versionNo int, dimensions []dpDatasetApiModels.Dimension, downloads *dpDatasetApiModels.DownloadList, alerts *[]dpDatasetApiModels.Alert) dpDatasetApiModels.Version {
	return dpDatasetApiModels.Version{
		ReleaseDate: fmt.Sprintf("01-0%d-2021", versionNo),
		Downloads:   downloads,
		Edition:     "2021",
		Version:     versionNo,
		Links: &dpDatasetApiModels.VersionLinks{
			Dataset: &dpDatasetApiModels.LinkObject{
				HRef: "http://localhost:22000/datasets/cantabular-1",
				ID:   "cantabular-1",
			},
		},
		Dimensions: dimensions,
		Alerts:     alerts,
	}
}

func getTestDimension(dimensionID string, isAreaType bool) dpDatasetApiModels.Dimension {
	return dpDatasetApiModels.Dimension{
		Description: fmt.Sprintf("A description for Dimension %s", dimensionID),
		Name:        dimensionID,
		ID:          dimensionID,
		Label:       fmt.Sprintf("Label %s", dimensionID),
		IsAreaType:  helpers.ToBoolPtr(isAreaType),
	}
}

func getTestFilterDimension(name string, isAreaType bool, options []string, categorisations int) sharedModel.FilterDimension {
	return sharedModel.FilterDimension{
		ModelDimension: filter.ModelDimension{
			Label:      fmt.Sprintf("Label %s", name),
			Options:    options,
			IsAreaType: helpers.ToBoolPtr(isAreaType),
			Name:       name,
			ID:         name,
		},
		OptionsCount:        len(options),
		CategorisationCount: categorisations,
	}
}

func buildTestFilterDimension(name string, isAreaType bool, optionCount int) sharedModel.FilterDimension {
	options := []string{}
	for i := 1; i <= optionCount; i++ {
		options = append(options, fmt.Sprintf("Label %d", i))
	}
	return getTestFilterDimension(name, isAreaType, options, 2)
}

func getTestDownloads(formats []string) *dpDatasetApiModels.DownloadList {
	downloadList := &dpDatasetApiModels.DownloadList{
		XLS:  &dpDatasetApiModels.DownloadObject{},
		XLSX: &dpDatasetApiModels.DownloadObject{},
		CSV:  &dpDatasetApiModels.DownloadObject{},
		TXT:  &dpDatasetApiModels.DownloadObject{},
		CSVW: &dpDatasetApiModels.DownloadObject{},
	}

	downloadObjects := downloadList.ExtensionsMapping()
	// Loop through the possible downloadobjects and add files to each format if requested
	for downloadObject, extension := range downloadObjects {
		if slices.Contains(formats, extension) {
			downloadObject.HRef = "https://mydomain.com/my-request"
			downloadObject.Size = "438290"
		}
	}
	return downloadList
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

func getTestDefaultDimensions() []dpDatasetApiModels.Dimension {
	dim1 := getTestDimension("1", true)
	dim1.QualityStatementText = "This is a quality notice statement"
	dim1.QualityStatementURL = "#"

	dim2 := getTestDimension("2", false)
	dim2.QualityStatementText = "This is another quality notice statement"
	dim2.QualityStatementURL = "#"

	dim3 := getTestDimension("3", false)
	dim3.Description = ""
	dim3.Name = "Only a name - I shouldn't map"

	return []dpDatasetApiModels.Dimension{dim1, dim2, dim3}
}

func getTestOptions(name string, count int) dpDatasetApiSdk.VersionDimensionOptionsList {
	items := []dpDatasetApiModels.PublicDimensionOption{}
	for i := 1; i <= count; i++ {
		items = append(items, dpDatasetApiModels.PublicDimensionOption{
			Name:   name,
			Label:  fmt.Sprintf("Label %d", i),
			Option: fmt.Sprintf("Label %d", i),
		})
	}

	return dpDatasetApiSdk.VersionDimensionOptionsList{
		Items: items,
	}
}

func getTestOptionsList() []dpDatasetApiSdk.VersionDimensionOptionsList {
	return []dpDatasetApiSdk.VersionDimensionOptionsList{
		getTestOptions("dim_1", 2),
	}
}

func getTestVersionOneDetails() dpDatasetApiModels.Version {
	return getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
}
