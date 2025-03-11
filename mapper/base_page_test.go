package mapper

import (
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dpRendererModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateBasePage(t *testing.T) {
	basePageModel := dpRendererModel.Page{}
	contacts := getTestContacts()
	isValidationError := false
	lang := "en"
	mockRequest := httptest.NewRequest("", "/", nil)
	relatedContent := getTestRelatedContent()
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	datasetDetails := getTestDatasetDetails(contacts, relatedContent)
	homepageContent := zebedee.HomepageContent{
		EmergencyBanner: emergencyBanner,
		ServiceMessage:  serviceMessage,
	}

	Convey("Test `UpdateBasePage` updates page attributes correctly default parameters", t, func() {

		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		// These parameters are set by default and are not dependent on conditional inputs
		So(basePageModel.BetaBannerEnabled, ShouldEqual, true)
		So(basePageModel.DatasetId, ShouldEqual, datasetDetails.ID)
		So(basePageModel.EmergencyBanner, ShouldEqual, mapEmergencyBanner(homepageContent.EmergencyBanner))
		So(basePageModel.Language, ShouldEqual, lang)
		So(basePageModel.Metadata.Description, ShouldEqual, datasetDetails.Description)
		So(basePageModel.Metadata.Title, ShouldEqual, datasetDetails.Title)
		So(basePageModel.ReleaseDate, ShouldEqual, "")
		So(basePageModel.ServiceMessage, ShouldEqual, serviceMessage)
		So(basePageModel.Type, ShouldEqual, datasetDetails.Type)
		So(basePageModel.URI, ShouldEqual, mockRequest.URL.Path)
	})

	Convey("Test `UpdateBasePage` does not update `Error` if `isValidationError` is `false", t, func() {
		isValidationError = false

		// Instantiation of `dpRendererModel.Page{}` sets `Error` to an empty struct
		expectedError := dpRendererModel.Error{
			Description: "",
			ErrorCode:   0,
			ErrorItems:  []dpRendererModel.ErrorItem(nil),
			Language:    "",
			Title:       "",
		}

		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		So(basePageModel.Error, ShouldEqual, expectedError)
	})

	Convey("Test `UpdateBasePage` updates `Error` if `isValidationError` is `true", t, func() {
		isValidationError = true

		// Error should be updated to show error details
		expectedError := dpRendererModel.Error{
			Description: "",
			ErrorCode:   0,
			ErrorItems: []dpRendererModel.ErrorItem{
				{
					Description: dpRendererModel.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
			Language: lang,
			Title:    datasetDetails.Title,
		}

		UpdateBasePage(&basePageModel, datasetDetails, homepageContent, isValidationError, lang, mockRequest)

		So(basePageModel.Error, ShouldEqual, expectedError)
	})
}
