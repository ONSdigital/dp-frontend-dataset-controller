package mapper

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusBasePage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
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
	versionThreeDetails := getTestVersionDetails(4, getTestDefaultDimensions(), getTestDownloads([]string{"xlsx"}), &[]dataset.Alert{})
	datasetOptions := getTestOptionsList()

	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("Census base maps correctly as version 1", t, func() {
		page := CreateCensusBasePage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
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
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 4)
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
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "/a/version/123", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.Version.ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeTrue)
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].VersionNumber, ShouldEqual, 2)
		So(page.Versions[0].ReleaseDate, ShouldEqual, versionTwoDetails.ReleaseDate)
		So(page.Versions[0].IsCurrentPage, ShouldBeTrue)
		So(page.Versions[0].Corrections[0].Reason, ShouldEqual, "This is a correction")
		So(page.DatasetLandingPage.LatestVersionURL, ShouldEqual, "/a/version/123")
	})

	Convey("IsCurrent returns false when request is for a different page", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/1", nil)
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].IsCurrentPage, ShouldBeFalse)
	})

	Convey("Versions history is in descending order", t, func() {
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Versions[0].VersionNumber, ShouldEqual, 3)
		So(page.Versions[1].VersionNumber, ShouldEqual, 2)
		So(page.Versions[2].VersionNumber, ShouldEqual, 1)
	})

	Convey("Given a census dataset landing page testing panels", t, func() {
		Convey("When there is more than one version", func() {
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					DisplayIcon: true,
					Body:        "New version",
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'other versions' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When there is one version", func() {
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{versionOneDetails}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When you are on the latest version", func() {
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionThreeDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When there a correction notice on the current version", func() {
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					DisplayIcon: true,
					Body:        "Correction notice",
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'correction notice' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When you are not on the latest version and a correction notice is on the current version", func() {
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					DisplayIcon: true,
					Body:        "Correction notice",
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
				{
					DisplayIcon: true,
					Body:        "New version",
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
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
			page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					DisplayIcon: true,
					Body:        "Important notice",
					CssClasses:  []string{"ons-u-mt-m", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'alert' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})
	})

	Convey("Validation error passed as true, error model should be populated", t, func() {
		req := httptest.NewRequest("", "/?f=get-data", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		mockErr := coreModel.Error{
			Title: datasetModel.Title,
			ErrorItems: []coreModel.ErrorItem{
				{
					Description: coreModel.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, true, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Error, ShouldResemble, mockErr)
	})

	Convey("Validation error passed as false, error title should be empty", t, func() {
		req := httptest.NewRequest("", "/?f=get-data&format=xlsx", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Error.Title, ShouldBeBlank)
	})

	Convey("Unknown get query request made, format selection error title should be empty", t, func() {
		req := httptest.NewRequest("", "/?f=blah-blah", nil)
		versionDetails := dataset.Version{
			Downloads: map[string]dataset.Download{
				"XLSX": {
					Size: "1234",
					URL:  "https://mydomain.com/my-request.xlsx",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Error.Title, ShouldBeBlank)
	})

	noContact := dataset.Contact{
		Telephone: "",
		Email:     "",
	}
	noContactDM := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			noContact,
		},
	}

	Convey("No contacts provided, contact section is not displayed", t, func() {
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, noContactDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.ContactDetails.Email, ShouldEqual, noContact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, noContact.Telephone)
		So(page.HasContactDetails, ShouldBeFalse)
	})

	oneContactDetail := dataset.Contact{
		Telephone: "123",
		Email:     "",
	}
	oneContactDetailDM := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			oneContactDetail,
		},
	}

	Convey("One contact detail provided, contact section is displayed", t, func() {
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, oneContactDetailDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.ContactDetails.Email, ShouldEqual, oneContactDetail.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, oneContactDetail.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
	})

	Convey("Dataset type is flexible, additional mapping is correct", t, func() {
		flexDm := dataset.DatasetDetails{
			Type: "cantabular_flexible_table",
			ID:   "test-flex",
		}
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, flexDm, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeFalse)
	})

	Convey("Dataset type is multivariate, additional mapping is correct", t, func() {
		mvd := dataset.DatasetDetails{
			Type: "cantabular_multivariate_table",
			ID:   "test-multi",
		}
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, mvd, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeTrue)
	})

	Convey("Config for multivariate=false, additional mapping is correct", t, func() {
		mvd := dataset.DatasetDetails{
			Type: "cantabular_multivariate_table",
			ID:   "test-multi",
		}
		page := CreateCensusDatasetLandingPage(false, context.Background(), req, pageModel, mvd, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeFalse)
		So(page.DatasetLandingPage.IsMultivariate, ShouldBeFalse)
	})
}
