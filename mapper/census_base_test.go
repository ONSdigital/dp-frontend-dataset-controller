package mapper

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/census"
	"github.com/ONSdigital/dp-renderer/v2/helper"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusBasePage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	pageModel := coreModel.Page{}
	contacts := getTestContacts()
	contact := contacts[0]
	relatedContent := getTestRelatedContent()
	datasetModel := getTestDatasetDetails(contacts, relatedContent)
	versionOneDetails := getTestVersionDetails(1, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), nil)
	versionTwoDetails := getTestVersionDetails(2, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}),
		&[]dataset.Alert{
			{
				Date:        "",
				Description: "This is a correction",
				Type:        "correction",
			},
		})
	versionThreeDetails := getTestVersionDetails(3, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), &[]dataset.Alert{})

	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()
	homepageContent := zebedee.HomepageContent{
		EmergencyBanner: emergencyBanner,
		ServiceMessage:  serviceMessage,
	}

	Convey("Census base maps correctly as version 1", t, func() {
		req := httptest.NewRequest("", "/", nil)
		// Call to `UpdateBasePage` adds base page metadata from request and dataset model
		UpdateBasePage(&pageModel, datasetModel, homepageContent, false, "en", req)
		page := CreateCensusBasePage(pageModel, datasetModel, versionOneDetails, []dataset.Version{versionOneDetails}, true)
		So(page.Type, ShouldEqual, datasetModel.Type)
		So(page.DatasetId, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.ReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.IsNationalStatistic, ShouldBeTrue)
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)

		So(page.DatasetLandingPage.RelatedContentItems[0].Title, ShouldEqual, relatedContent[0].Title)
		So(page.DatasetLandingPage.RelatedContentItems[1].Title, ShouldEqual, relatedContent[1].Title)
		So(page.Page.ServiceMessage, ShouldEqual, serviceMessage)
		So(page.Page.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
		So(page.Page.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
		So(page.Page.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
		So(page.Page.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
		So(page.Page.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)

		So(page.ShowCensusBranding, ShouldBeTrue)
	})

	Convey("Release date and hasOtherVersions is mapped correctly when v2 of Census DLP dataset is loaded", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/2", nil)
		// Call to `UpdateBasePage` adds base page metadata from request and dataset model
		UpdateBasePage(&pageModel, datasetModel, homepageContent, false, "en", req)
		page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails}, true)
		So(page.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.Version.ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeTrue)
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].VersionNumber, ShouldEqual, 2)
		So(page.Versions[0].ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.Versions[0].IsCurrentPage, ShouldBeTrue)
		So(page.Versions[0].Corrections[0].Reason, ShouldEqual, "This is a correction")
		So(page.DatasetLandingPage.LatestVersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
	})

	Convey("IsCurrent returns false when request is for a different page", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/1", nil)
		// Call to `UpdateBasePage` adds base page metadata from request and dataset model
		UpdateBasePage(&pageModel, datasetModel, homepageContent, false, "en", req)
		page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails}, true)
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].IsCurrentPage, ShouldBeFalse)
	})

	Convey("Versions history is in descending order", t, func() {
		page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, true)
		So(page.Versions[0].VersionNumber, ShouldEqual, 3)
		So(page.Versions[1].VersionNumber, ShouldEqual, 2)
		So(page.Versions[2].VersionNumber, ShouldEqual, 1)
	})

	Convey("Given a census dataset landing page testing panels", t, func() {
		Convey("When there is more than one version", func() {
			page := CreateCensusBasePage(pageModel, datasetModel, versionOneDetails, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, true)
			mockPanel := []census.Panel{
				{
					DisplayIcon: true,
					Body:        []string{"New version"},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'other versions' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When there is one version", func() {
			page := CreateCensusBasePage(pageModel, datasetModel, versionOneDetails, []dataset.Version{versionOneDetails}, true)
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When you are on the latest version", func() {
			page := CreateCensusBasePage(pageModel, datasetModel, versionThreeDetails, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, true)
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When there a correction notice on the current version", func() {
			page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails}, true)
			mockPanel := []census.Panel{
				{
					DisplayIcon: true,
					Body:        []string{"Correction notice"},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'correction notice' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When you are not on the latest version and a correction notice is on the current version", func() {
			page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, true)
			mockPanel := []census.Panel{
				{
					DisplayIcon: true,
					Body:        []string{"Correction notice"},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
				{
					DisplayIcon: true,
					Body:        []string{"New version"},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'correction notice' and 'other versions' panels are displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 2)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When there is an alert on the current version", func() {
			versionTwoDetails.Alerts = &[]dataset.Alert{
				{
					Description: "Important notice",
					Type:        "alert",
				},
			}
			page := CreateCensusBasePage(pageModel, datasetModel, versionTwoDetails, []dataset.Version{versionOneDetails, versionTwoDetails}, true)
			mockPanel := []census.Panel{
				{
					DisplayIcon: true,
					Body:        []string{"Important notice"},
					CSSClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'alert' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})
	})

	Convey("Unknown get query request made, format selection error title should be empty", t, func() {
		page := CreateCensusBasePage(pageModel, datasetModel, versionOneDetails, []dataset.Version{}, true)
		So(page.Error.Title, ShouldBeBlank)
	})

	noContacts := []dataset.Contact{
		{
			Telephone: "",
			Email:     "",
		}}
	noContactDM := getTestDatasetDetails(noContacts, relatedContent)

	Convey("No contacts provided, contact section is not displayed", t, func() {
		page := CreateCensusBasePage(pageModel, noContactDM, versionOneDetails, []dataset.Version{}, true)
		So(page.ContactDetails.Email, ShouldEqual, "")
		So(page.ContactDetails.Telephone, ShouldEqual, "")
		So(page.HasContactDetails, ShouldBeFalse)
	})

	oneContactDetail := []dataset.Contact{
		{
			Telephone: "123",
			Email:     "",
		}}
	oneContactDetailDM := getTestDatasetDetails(oneContactDetail, relatedContent)

	Convey("One contact detail provided, contact section is displayed", t, func() {
		page := CreateCensusBasePage(pageModel, oneContactDetailDM, versionOneDetails, []dataset.Version{}, true)
		So(page.ContactDetails.Email, ShouldEqual, oneContactDetail[0].Email)
		So(page.ContactDetails.Telephone, ShouldEqual, oneContactDetail[0].Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
	})

	Convey("Dataset type is flexible, additional mapping is correct", t, func() {
		flexDm := dataset.DatasetDetails{
			Type: "cantabular_flexible_table",
			ID:   "test-flex",
		}
		page := CreateCensusBasePage(pageModel, flexDm, versionOneDetails, []dataset.Version{}, true)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeFalse)
	})

	Convey("Dataset type is multivariate, additional mapping is correct", t, func() {
		mvd := dataset.DatasetDetails{
			Type: "cantabular_multivariate_table",
			ID:   "test-multi",
		}
		page := CreateCensusBasePage(pageModel, mvd, versionOneDetails, []dataset.Version{}, true)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeTrue)
	})

	Convey("Config for multivariate=false, additional mapping is correct", t, func() {
		mvd := dataset.DatasetDetails{
			Type: "cantabular_multivariate_table",
			ID:   "test-multi",
		}
		page := CreateCensusBasePage(pageModel, mvd, versionOneDetails, []dataset.Version{}, false)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeFalse)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeFalse)
	})
}
