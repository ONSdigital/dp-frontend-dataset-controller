package mapper

import (
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-dataset-api/models"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticOverviewPage(t *testing.T) {
	basePage := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	mockRequest := httptest.NewRequest("", "/", nil)

	datasetModel := getTestDatasetDetails(contacts, relatedContent)

	version := models.Version{
		ReleaseDate: "2025-01-26T07:00:00.000Z",
		UsageNotes: &[]models.UsageNote{
			{
				Note: `To assist individuals in understanding how the rise in inflation affects 
					their expenditure, we have published a personal inflation calculator. It enables 
					consumers to enter the amounts they spend against different categories, and the 
					calculator will provide an estimate of their personal inflation based on those 
					spending patterns.`,
				Title: "Interactive Personal Inflation Calculator",
			},
		},
		Version: 1,
	}

	Convey("Test mapper returns static page with correct attributes", t, func() {
		staticPage := CreateStaticOverviewPage(mockRequest, basePage, datasetModel, version)
	})
}

// func TestCreateCustomDataset(t *testing.T) {
// 	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
// 	req := httptest.NewRequest("", "/", nil)
// 	pageModel := coreModel.Page{}
// 	serviceMessage := getTestServiceMessage()
// 	emergencyBanner := getTestEmergencyBanner()

// 	Convey("Given simple content for a page", t, func() {
