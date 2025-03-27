package mapper

import (
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateStaticOverviewPage(t *testing.T) {
	basePage := coreModel.Page{}
	contacts := getTestContacts()
	relatedContent := getTestRelatedContent()
	mockRequest := httptest.NewRequest("", "/", nil)

	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	version := model.Version{
		Title: "Consumer price inflation tables",
		// Title         string                 `json:"title"`
		// Description   string                 `json:"description"`
		// URL           string                 `json:"url"`
		// ReleaseDate   string                 `json:"release_date"`
		// NextRelease   string                 `json:"next_release"`
		// Downloads     []Download             `json:"downloads"`
		// Edition       string                 `json:"edition"`
		// Version       string                 `json:"version"`
		// Contact       contact.Details        `json:"contact"`
		// IsCurrentPage bool                   `json:"is_current"`
		// VersionURL    string                 `json:"version_url"`
		// Superseded    string                 `json:"superseded"`
		// VersionNumber int                    `json:"version_number"`
		// Date          string                 `json:"date"`
		// Corrections   []Correction           `json:"correction"`
		// FilterURL     string                 `json:"filter_url"`
		// IsLatest      bool                   `json:"is_latest"`
		// Distributions *[]models.Distribution `json:"distributions,omitempty"`
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
