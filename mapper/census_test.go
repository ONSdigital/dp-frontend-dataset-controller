package mapper

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper/mocks"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/helper"
	coreModel "github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusDatasetLandingPage(t *testing.T) {
	helper.InitialiseLocalisationsHelper(mocks.MockAssetFunction)
	req := httptest.NewRequest("", "/", nil)
	pageModel := coreModel.Page{}
	contact := dataset.Contact{
		Telephone: "01232 123 123",
		Email:     "hello@testing.com",
	}
	relatedContent := []dataset.GeneralDetails{
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

	datasetModel := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			contact,
		},
		ID:                "12345",
		Description:       "An interesting test description \n with a line break",
		Title:             "Test title",
		Type:              "cantabular",
		NationalStatistic: true,
		Survey:            "census",
		RelatedContent:    &relatedContent,
	}

	versionOneDetails := dataset.Version{
		ReleaseDate: "01-01-2021",
		Downloads: map[string]dataset.Download{
			"XLSX": {
				Size: "438290",
				URL:  "https://mydomain.com/my-request",
			},
		},
		Edition: "2021",
		Version: 1,
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cantabular-1",
				ID:  "cantabular-1",
			},
		},
		Dimensions: []dataset.VersionDimension{
			{
				Description:          "A description on one line",
				Name:                 "Dimension 1",
				ID:                   "dim_1",
				Label:                "Label 1",
				IsAreaType:           helpers.ToBoolPtr(true),
				QualityStatementText: "This is a quality notice statement",
				QualityStatementURL:  "#",
			},
			{
				Description:          "A description on one line \n Then a line break",
				Name:                 "Dimension 2",
				Label:                "Label 2",
				ID:                   "dim_2",
				QualityStatementText: "This is another quality notice statement",
				QualityStatementURL:  "#",
			},
			{
				Description: "",
				Name:        "Only a name - I shouldn't map",
				Label:       "Label 3",
				ID:          "dim_3",
			},
		},
	}

	versionTwoDetails := dataset.Version{
		ReleaseDate: "15-02-2021",
		Version:     2,
		Edition:     "2021",
		Links: dataset.Links{
			Dataset: dataset.Link{
				URL: "http://localhost:22000/datasets/cantabular-1",
				ID:  "cantabular-1",
			},
		},
		Alerts: &[]dataset.Alert{
			{
				Date:        "",
				Description: "This is a correction",
				Type:        "correction",
			},
		},
	}

	versionThreeDetails := versionTwoDetails
	versionThreeDetails.Version = 3
	versionThreeDetails.Alerts = &[]dataset.Alert{}

	datasetOptions := []dataset.Options{
		{
			Items: []dataset.Option{
				{
					DimensionID: "dim_1",
					Option:      "option 1",
				},
				{
					DimensionID: "dim_1",
					Option:      "option 2",
				},
			},
		},
	}

	filterOutput := map[string]filter.Download{
		"csv": {
			Size: "12345",
			URL:  "https://mydomain.com/my-request",
		},
		"xls": {
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
	}

	fDims := []sharedModel.FilterDimension{
		{
			ModelDimension: filter.ModelDimension{
				Label:      "A label",
				Options:    []string{"An option", "and another"},
				IsAreaType: helpers.ToBoolPtr(true),
				Name:       "Geography",
			},
			OptionsCount: 2,
		},
	}

	serviceMessage := getTestServiceMessage()
	emergencyBanner := getTestEmergencyBanner()

	Convey("Census dataset landing page maps correctly as version 1", t, func() {
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)
		So(page.Type, ShouldEqual, datasetModel.Type)
		So(page.DatasetId, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.ReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		So(page.Version.Downloads[0].Size, ShouldEqual, "438290")
		So(page.Version.Downloads[0].Extension, ShouldEqual, "xlsx")
		So(page.Version.Downloads[0].URI, ShouldEqual, "https://mydomain.com/my-request")
		So(page.Version.Downloads, ShouldHaveLength, 1)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.IsNationalStatistic, ShouldBeTrue)
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Coverage")
		So(page.Collapsible.CollapsibleItems[2].Subheading, ShouldEqual, versionOneDetails.Dimensions[0].Label)
		So(page.Collapsible.CollapsibleItems[2].Content[0], ShouldEqual, versionOneDetails.Dimensions[0].Description)
		So(page.Collapsible.CollapsibleItems[3].Subheading, ShouldEqual, versionOneDetails.Dimensions[1].Label)
		So(page.Collapsible.CollapsibleItems[3].Content, ShouldResemble, strings.Split(versionOneDetails.Dimensions[1].Description, "\n"))
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 4)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions, ShouldHaveLength, 2) // coverage is inserted
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Coverage")
		So(page.DatasetLandingPage.Dimensions[1].Name, ShouldEqual, "coverage")
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeFalse)
		So(page.DatasetLandingPage.RelatedContentItems[0].Title, ShouldEqual, relatedContent[0].Title)
		So(page.DatasetLandingPage.RelatedContentItems[1].Title, ShouldEqual, relatedContent[1].Title)
		So(page.Page.ServiceMessage, ShouldEqual, serviceMessage)
		So(page.Page.EmergencyBanner.Type, ShouldEqual, strings.Replace(emergencyBanner.Type, "_", "-", -1))
		So(page.Page.EmergencyBanner.Title, ShouldEqual, emergencyBanner.Title)
		So(page.Page.EmergencyBanner.Description, ShouldEqual, emergencyBanner.Description)
		So(page.Page.EmergencyBanner.URI, ShouldEqual, emergencyBanner.URI)
		So(page.Page.EmergencyBanner.LinkText, ShouldEqual, emergencyBanner.LinkText)
		So(page.SearchNoIndexEnabled, ShouldBeFalse)
		So(page.ShowCensusBranding, ShouldBeTrue)
	})

	Convey("Census dataset landing page maps correctly with filter output", t, func() {
		datasetModel.Type = "flexible"
		page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, true, true, filterOutput, fDims, serviceMessage, emergencyBanner)
		So(page.Type, ShouldEqual, fmt.Sprintf("%s_filter_output", datasetModel.Type))
		So(page.DatasetId, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.ReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		// Downloads are mapped from filterOutput
		So(page.Version.Downloads, ShouldHaveLength, 4)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, "Area type")
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, "Coverage")
		// TODO: Removing test coverage until API is created
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 2)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, fDims[0].Label)
		So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, fDims[0].Options)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[0].Name, ShouldEqual, fDims[0].Name)
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, fDims[0].Options)
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)
		So(page.DatasetLandingPage.RelatedContentItems[0].Title, ShouldEqual, relatedContent[0].Title)
		So(page.DatasetLandingPage.RelatedContentItems[1].Title, ShouldEqual, relatedContent[1].Title)
		So(page.SearchNoIndexEnabled, ShouldBeTrue)
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

		Convey("When there is a quality notice on a dimension", func() {
			datasetOptions := []dataset.Options{
				{
					Items: []dataset.Option{
						{
							DimensionID: "Dimension 1",
							Label:       "Label 1",
						},
					},
				},
				{
					Items: []dataset.Option{
						{
							DimensionID: "Dimension 2",
							Label:       "Label 1",
						},
					},
				},
			}
			page := CreateCensusDatasetLandingPage(
				true,
				context.Background(),
				req,
				pageModel,
				datasetModel,
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
				false,
				false,
				map[string]filter.Download{},
				[]sharedModel.FilterDimension{},
				serviceMessage,
				emergencyBanner)

			mockPanel := []datasetLandingPageCensus.Panel{
				{
					Body:       "<p>This is a quality notice statement</p>Read more about this",
					CssClasses: []string{"ons-u-mt-no"},
				},
				{
					Body:       "<p>This is another quality notice statement</p>Read more about this",
					CssClasses: []string{"ons-u-mt-no", "ons-u-mb-l"},
				},
			}
			Convey("Then the 'quality notice' panel is displayed", func() {
				So(page.DatasetLandingPage.QualityStatements, ShouldHaveLength, 2)
				So(page.DatasetLandingPage.QualityStatements, ShouldResemble, mockPanel)
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

	Convey("Downloads are properly mapped", t, func() {
		Convey("On version", func() {
			Convey("Where all four downloads exist", func() {
				page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: map[string]dataset.Download{
					"csv": {
						Size: "12345",
						URL:  "https://mydomain.com/my-request",
					},
					"xls": {
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
				}}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)

				Convey("HasDownloads set to true when downloads are greater than three or more", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
				})

				Convey("GetDataXLSXInfo set to false", func() {
					So(page.DatasetLandingPage.ShowXLSXInfo, ShouldBeFalse)
				})

				Convey("Downloads are sorted by fixed extension order", func() {
					So(page.Version.Downloads[0].Extension, ShouldEqual, "xls")
					So(page.Version.Downloads[1].Extension, ShouldEqual, "csv")
					So(page.Version.Downloads[2].Extension, ShouldEqual, "txt")
					So(page.Version.Downloads[3].Extension, ShouldEqual, "csvw")
				})
			})
			Convey("Where excel download is missing", func() {
				page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: map[string]dataset.Download{
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
				}}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)

				Convey("HasDownloads set to true when downloads are greater than three or more", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
				})

				Convey("Downloads are sorted by fixed extension order", func() {
					So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
					So(page.Version.Downloads[1].Extension, ShouldEqual, "txt")
					So(page.Version.Downloads[2].Extension, ShouldEqual, "csvw")
				})
			})
			Convey("Where no downloads exist", func() {
				page := CreateCensusDatasetLandingPage(true, context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: nil}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, map[string]filter.Download{}, []sharedModel.FilterDimension{}, serviceMessage, emergencyBanner)

				Convey("HasDownloads set to false", func() {
					So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
				})
			})
		})

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

	Convey("given a dimension to truncate on census dataset landing page", t, func() {
		datasetOptions := []dataset.Options{
			{
				Items: []dataset.Option{
					{
						DimensionID: "Dimension 1",
						Label:       "Label 1",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 2",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 3",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 4",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 5",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 6",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 7",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 8",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 9",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 10",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 11",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 12",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 13",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 14",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 15",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 16",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 17",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 18",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 19",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 20",
					},
					{
						DimensionID: "Dimension 1",
						Label:       "Label 21",
					},
				},
				TotalCount: 20,
			},
			{
				Items: []dataset.Option{
					{
						DimensionID: "Dimension 2",
						Label:       "Label 1",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 2",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 3",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 4",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 5",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 6",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 7",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 8",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 9",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 10",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 11",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 12",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 13",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 14",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 15",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 16",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 17",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 18",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 19",
					},
					{
						DimensionID: "Dimension 2",
						Label:       "Label 20",
					},
				},
				TotalCount: 20,
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
				false,
				false,
				map[string]filter.Download{},
				[]sharedModel.FilterDimension{},
				serviceMessage,
				emergencyBanner)
			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 19", "Label 20", "Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/?showAll=dim_1#dim_1")
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
				false,
				false,
				map[string]filter.Download{},
				[]sharedModel.FilterDimension{},
				serviceMessage,
				emergencyBanner)

			Convey("then the dimension is no longer truncated", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, datasetOptions[0].TotalCount)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].IsTruncated, ShouldBeFalse)
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
			})

			Convey("then other truncated dimensions are persisted", func() {
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldHaveLength, 21)
				So(p.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
					"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
					"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
					"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
					"Label 21",
				})
				So(p.DatasetLandingPage.Dimensions[0].TruncateLink, ShouldEqual, "/#dim_1")
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, datasetOptions[1].TotalCount)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 9", "Label 10", "Label 11",
					"Label 18", "Label 19", "Label 20",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
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
