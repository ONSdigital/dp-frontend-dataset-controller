package mapper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-design-system-go/helper"
	core "github.com/ONSdigital/dis-design-system-go/model"
	"github.com/ONSdigital/dp-api-clients-go/v2/population"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCustomDataset(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", http.NoBody)
	pageModel := core.Page{}
	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("Given simple content for a page", t, func() {
		populationTypes := []population.PopulationType{
			{
				Name:        "Name 1",
				Label:       "Label 1",
				Description: "Description 1",
			},
			{
				Name:        "Name 2",
				Label:       "Label 2",
				Description: "Description 2",
			},
			{
				Name:        "Name 3",
				Label:       "Label 3",
				Description: "Description 3",
			},
		}

		Convey("When we build a census landing page", func() {
			page := CreateCustomDatasetPage(req, pageModel, populationTypes, "", serviceMessage, emergencyBanner)

			Convey("Then population types should be mapped correctly", func() {
				So(page.CreateCustomDatasetPage.PopulationTypes[0].Name, ShouldEqual, "Name 1")
				So(page.CreateCustomDatasetPage.PopulationTypes[1].Name, ShouldEqual, "Name 2")
				So(page.CreateCustomDatasetPage.PopulationTypes[2].Name, ShouldEqual, "Name 3")

				So(page.CreateCustomDatasetPage.PopulationTypes[0].Label, ShouldEqual, "Label 1")
				So(page.CreateCustomDatasetPage.PopulationTypes[0].Description, ShouldEqual, "Description 1")
			})
		})
	})
}
