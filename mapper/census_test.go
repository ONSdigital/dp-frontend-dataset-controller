package mapper

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	"github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateCensusDatasetLandingPage(t *testing.T) {
	req := httptest.NewRequest("", "/", nil)
	pageModel := model.Page{}
	contact := dataset.Contact{
		Telephone: "01232 123 123",
		Email:     "hello@testing.com",
	}
	datasetModel := dataset.DatasetDetails{
		Contacts: &[]dataset.Contact{
			contact,
		},
		ID:          "12345",
		Description: "An interesting test description \n with a line break",
		Title:       "Test title",
		Type:        "cantabular",
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
				Description: "A description on one line",
				Name:        "Dimension 1",
				ID:          "dim_1",
				IsAreaType:  helpers.ToBoolPtr(true),
			},
			{
				Description: "A description on one line \n Then a line break",
				Name:        "Dimension 2",
				ID:          "dim_2",
			},
			{
				Description: "",
				Name:        "Only a name - I shouldn't map",
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

	filterOutput := filter.Model{
		Dimensions: []filter.ModelDimension{
			{
				Label:      "A label",
				Options:    []string{"An option", "and another"},
				IsAreaType: helpers.ToBoolPtr(true),
				Name:       "Geography",
			},
		},
		Downloads: map[string]filter.Download{
			"CSV": {
				Size: "12345",
				URL:  "https://mydomain.com/my-request",
			},
		},
		FilterID: "1234-5678",
	}

	Convey("Census dataset landing page maps correctly as version 1", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, false, false, filter.Model{})
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
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, versionOneDetails.Dimensions[0].Name)
		So(page.Collapsible.CollapsibleItems[0].Content[0], ShouldEqual, versionOneDetails.Dimensions[0].Description)
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, versionOneDetails.Dimensions[1].Name)
		So(page.Collapsible.CollapsibleItems[1].Content, ShouldResemble, strings.Split(versionOneDetails.Dimensions[1].Description, "\n"))
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 2)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions, ShouldHaveLength, 2) // coverage is inserted
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Title, ShouldEqual, "Coverage")
		So(page.DatasetLandingPage.Dimensions[1].Name, ShouldEqual, "coverage")
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeFalse)
	})

	Convey("Census dataset landing page maps correctly with filter output", t, func() {
		datasetModel.Type = "flexible"
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, "", false, []dataset.Version{versionOneDetails}, 1, "/a/version/1", "", []string{}, 50, false, true, true, filterOutput)
		So(page.Type, ShouldEqual, fmt.Sprintf("%s_filter_output", datasetModel.Type))
		So(page.DatasetId, ShouldEqual, datasetModel.ID)
		So(page.Version.ReleaseDate, ShouldEqual, versionOneDetails.ReleaseDate)
		So(page.ReleaseDate, ShouldEqual, page.Version.ReleaseDate)
		So(page.DatasetLandingPage.HasOtherVersions, ShouldBeFalse)
		// Downloads are mapped from filterOutput
		So(page.Version.Downloads[0].Size, ShouldEqual, "12345")
		So(page.Version.Downloads[0].Extension, ShouldEqual, "csv")
		So(page.Version.Downloads[0].URI, ShouldEqual, "https://mydomain.com/my-request")
		So(page.Version.Downloads, ShouldHaveLength, 1)
		So(page.Metadata.Title, ShouldEqual, datasetModel.Title)
		So(page.Metadata.Description, ShouldEqual, datasetModel.Description)
		So(page.DatasetLandingPage.Description, ShouldResemble, strings.Split(datasetModel.Description, "\n"))
		So(page.ContactDetails.Name, ShouldEqual, contact.Name)
		So(page.ContactDetails.Email, ShouldEqual, contact.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, contact.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
		So(page.DatasetLandingPage.LatestVersionURL, ShouldBeBlank)
		So(page.Collapsible.CollapsibleItems[0].Subheading, ShouldEqual, versionOneDetails.Dimensions[0].Name)
		So(page.Collapsible.CollapsibleItems[0].Content[0], ShouldEqual, versionOneDetails.Dimensions[0].Description)
		So(page.Collapsible.CollapsibleItems[1].Subheading, ShouldEqual, versionOneDetails.Dimensions[1].Name)
		So(page.Collapsible.CollapsibleItems[1].Content, ShouldResemble, strings.Split(versionOneDetails.Dimensions[1].Description, "\n"))
		So(page.Collapsible.CollapsibleItems, ShouldHaveLength, 2)
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeFalse)
		So(page.DatasetLandingPage.Dimensions[0].Title, ShouldEqual, filterOutput.Dimensions[0].Label)
		So(page.DatasetLandingPage.Dimensions[0].Values, ShouldResemble, filterOutput.Dimensions[0].Options)
		So(page.DatasetLandingPage.Dimensions[0].ShowChange, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[0].ChangeURL, ShouldEqual, "/filters/1234-5678/dimensions/geography")
		So(page.DatasetLandingPage.Dimensions[1].IsCoverage, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].Values, ShouldResemble, filterOutput.Dimensions[0].Options)
		So(page.DatasetLandingPage.Dimensions[1].ShowChange, ShouldBeTrue)
		So(page.DatasetLandingPage.Dimensions[1].ChangeURL, ShouldEqual, "/filters/1234-5678/dimensions/geography/coverage")
	})

	Convey("Release date and hasOtherVersions is mapped correctly when v2 of Census DLP dataset is loaded", t, func() {
		req := httptest.NewRequest("", "/datasets/cantabular-1/editions/2021/versions/2", nil)
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "/a/version/123", "", []string{}, 50, false, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Versions[0].VersionURL, ShouldEqual, "/datasets/cantabular-1/editions/2021/versions/2")
		So(page.Versions[0].IsCurrentPage, ShouldBeFalse)
	})

	Convey("Versions history is in descending order", t, func() {
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.Versions[0].VersionNumber, ShouldEqual, 3)
		So(page.Versions[1].VersionNumber, ShouldEqual, 2)
		So(page.Versions[2].VersionNumber, ShouldEqual, 1)
	})

	Convey("Given a census dataset landing page testing panels", t, func() {
		Convey("When there is more than one version", func() {
			page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					IsCorrection: false,
				},
			}
			Convey("Then the 'other versions' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When there is one version", func() {
			page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionOneDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{versionOneDetails}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When you are on the latest version", func() {
			page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionThreeDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
			Convey("Then the 'other versions' panel is not displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldBeEmpty)
			})
		})

		Convey("When there a correction notice on the current version", func() {
			page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails}, 2, "", "", []string{}, 50, false, false, false, filter.Model{})
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					IsCorrection: true,
				},
			}
			Convey("Then the 'correction notice' panel is displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 1)
				So(page.DatasetLandingPage.Panels, ShouldResemble, mockPanel)
			})
		})

		Convey("When you are not on the latest version and a correction notice is on the current version", func() {
			page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionTwoDetails, datasetOptions, versionOneDetails.ReleaseDate, true, []dataset.Version{versionOneDetails, versionTwoDetails, versionThreeDetails}, 3, "", "", []string{}, 50, false, false, false, filter.Model{})
			mockPanel := []datasetLandingPageCensus.Panel{
				{
					IsCorrection: true,
				},
				{
					IsCorrection: false,
				},
			}
			Convey("Then the 'correction notice' and 'other versions' panels are displayed", func() {
				So(page.DatasetLandingPage.Panels, ShouldHaveLength, 2)
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
		mockErr := model.Error{
			Title: datasetModel.Title,
			ErrorItems: []model.ErrorItem{
				{
					Description: model.Localisation{
						LocaleKey: "GetDataValidationError",
						Plural:    1,
					},
					URL: "#select-format-error",
				},
			},
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, true, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, datasetModel, versionDetails, datasetOptions, versionOneDetails.ReleaseDate, false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, noContactDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
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
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.ContactDetails.Email, ShouldEqual, oneContactDetail.Email)
		So(page.ContactDetails.Telephone, ShouldEqual, oneContactDetail.Telephone)
		So(page.HasContactDetails, ShouldBeTrue)
	})

	Convey("Dataset type is flexible, additional mapping is correct", t, func() {
		flexDm := dataset.DatasetDetails{
			Type: "cantabular_flexible_table",
			ID:   "test-flex",
		}
		page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, flexDm, versionOneDetails, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
		So(page.DatasetLandingPage.IsFlexibleForm, ShouldBeTrue)
		So(page.DatasetLandingPage.FormAction, ShouldEqual, fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/filter-flex", flexDm.ID, versionOneDetails.Edition, strconv.Itoa(versionOneDetails.Version)))
	})

	Convey("Test HasDownloads", t, func() {
		Convey("On version", func() {
			Convey("HasDownloads set to true when downloads are greater than zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: map[string]dataset.Download{
					"XLSX": {
						Size: "1234",
						URL:  "https://mydomain.com/my-request.xlsx",
					},
				}}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})
			Convey("HasDownloads set to false when downloads are zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(), req, pageModel, oneContactDetailDM, dataset.Version{Downloads: nil}, datasetOptions, "", false, []dataset.Version{}, 1, "", "", []string{}, 50, false, false, false, filter.Model{})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
			})
		})
		Convey("On filterOutput", func() {
			Convey("HasDownloads set to true when downloads are greater than zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(),
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
					filter.Model{
						Dimensions: []filter.ModelDimension{
							{
								Name:       "Area 1",
								IsAreaType: helpers.ToBoolPtr(true),
								Options:    []string{"one", "two", "three"},
							},
						},
						Downloads: map[string]filter.Download{
							"XLSX": {
								Size: "1234",
								URL:  "https://mydomain.com/my-request.xlsx",
							},
						}})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeTrue)
			})
			Convey("HasDownloads set to false when downloads are zero", func() {
				page := CreateCensusDatasetLandingPage(context.Background(),
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
					filter.Model{
						Dimensions: []filter.ModelDimension{
							{
								Name:       "Area 1",
								IsAreaType: helpers.ToBoolPtr(true),
								Options:    []string{"one", "two", "three"},
							},
						},
						Downloads: nil,
					})
				So(page.DatasetLandingPage.HasDownloads, ShouldBeFalse)
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
				filter.Model{Downloads: nil})
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
				filter.Model{Downloads: nil})

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
		filterDims := filter.Model{
			Dimensions: []filter.ModelDimension{
				{
					Label: "Dimension 1",
					ID:    "dim_1",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4", "Label 5",
						"Label 6", "Label 7", "Label 8", "Label 9", "Label 10",
						"Label 11", "Label 12", "Label 13", "Label 14", "Label 15",
						"Label 16", "Label 17", "Label 18", "Label 19", "Label 20",
						"Label 21",
					},
				},
				{
					Label: "Dimension 2",
					ID:    "dim_2",
					Options: []string{
						"Label 1", "Label 2", "Label 3", "Label 4",
						"Label 5", "Label 6", "Label 7", "Label 8",
						"Label 9", "Label 10", "Label 11", "Label 12",
					},
				},
			},
		}

		Convey("when valid parameters are provided", func() {
			p := CreateCensusDatasetLandingPage(
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
				filterDims)
			Convey("then the list should be truncated to show the first, middle, and last three values", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, 21)
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
				filterDims)

			Convey("then the dimension is no longer truncated", func() {
				So(p.DatasetLandingPage.Dimensions[0].TotalItems, ShouldEqual, 21)
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
				So(p.DatasetLandingPage.Dimensions[2].TotalItems, ShouldEqual, 12)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldHaveLength, 9)
				So(p.DatasetLandingPage.Dimensions[2].Values, ShouldResemble, []string{
					"Label 1", "Label 2", "Label 3",
					"Label 5", "Label 6", "Label 7",
					"Label 10", "Label 11", "Label 12",
				})
				So(p.DatasetLandingPage.Dimensions[2].IsTruncated, ShouldBeTrue)
				So(p.DatasetLandingPage.Dimensions[2].TruncateLink, ShouldEqual, "/?showAll=dim_2#dim_2")
			})
		})
	})
}
