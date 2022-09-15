package mapper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	sharedModel "github.com/ONSdigital/dp-frontend-dataset-controller/model"
	"github.com/ONSdigital/dp-frontend-dataset-controller/model/datasetLandingPageCensus"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

// Constants...
const (
	CorrectionAlertType = "correction"
	queryStrKey         = "showAll"
	Coverage            = "Coverage"
	FilterOutput        = "_filter_output"
)

// CreateCensusDatasetLandingPage creates a census-landing page based on api model responses
func CreateCensusDatasetLandingPage(ctx context.Context, req *http.Request, basePage coreModel.Page, d dataset.DatasetDetails, version dataset.Version, opts []dataset.Options, initialVersionReleaseDate string, hasOtherVersions bool, allVersions []dataset.Version, latestVersionNumber int, latestVersionURL, lang string, queryStrValues []string, maxNumberOfOptions int, isValidationError, isFilterOutput, hasNoAreaOptions bool, filter filter.Model) datasetLandingPageCensus.Page {
	p := datasetLandingPageCensus.Page{
		Page: basePage,
	}

	MapCookiePreferences(req, &p.Page.CookiesPreferencesSet, &p.Page.CookiesPolicy)

	p.Type = d.Type
	if isFilterOutput {
		p.Type += FilterOutput
	}
	p.Language = lang
	p.URI = req.URL.Path
	p.DatasetId = d.ID
	p.Version.ReleaseDate = version.ReleaseDate
	if initialVersionReleaseDate == "" {
		p.ReleaseDate = p.Version.ReleaseDate
	} else {
		p.ReleaseDate = initialVersionReleaseDate
	}

	if version.Alerts != nil {
		for _, alert := range *version.Alerts {
			if alert.Type == CorrectionAlertType {
				p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, datasetLandingPageCensus.Panel{
					IsCorrection: true,
				})
				break
			}
		}
	}
	p.DatasetLandingPage.HasOtherVersions = hasOtherVersions
	p.Metadata.Title = d.Title
	p.Metadata.Description = d.Description
	var isFlex bool
	if strings.Contains(d.Type, "flex") {
		isFlex = true
		p.DatasetLandingPage.IsFlexibleForm = true
		p.DatasetLandingPage.FormAction = fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/filter-flex", d.ID, version.Edition, strconv.Itoa(version.Version))
	}

	if isFilterOutput {
		for ext, download := range filter.Downloads {
			p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
				Extension: strings.ToLower(ext),
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	} else {
		for ext, download := range version.Downloads {
			p.Version.Downloads = append(p.Version.Downloads, sharedModel.Download{
				Extension: strings.ToLower(ext),
				Size:      download.Size,
				URI:       download.URL,
			})
		}
	}

	if d.Contacts != nil && len(*d.Contacts) > 0 {
		contacts := *d.Contacts
		if contacts[0].Telephone != "" {
			p.ContactDetails.Telephone = contacts[0].Telephone
			p.HasContactDetails = true
		}
		if contacts[0].Email != "" {
			p.ContactDetails.Email = contacts[0].Email
			p.HasContactDetails = true
		}
	}

	p.DatasetLandingPage.Description = strings.Split(d.Description, "\n")

	var collapsibleContentItems []coreModel.CollapsibleItem

	for _, dims := range version.Dimensions {
		if dims.Description != "" {
			var collapsibleContent coreModel.CollapsibleItem
			collapsibleContent.Subheading = dims.Name
			collapsibleContent.Content = strings.Split(dims.Description, "\n")
			collapsibleContentItems = append(collapsibleContentItems, collapsibleContent)
		}
	}

	if len(collapsibleContentItems) > 0 {
		p.Collapsible = coreModel.Collapsible{
			Title: coreModel.Localisation{
				LocaleKey: "VariablesExplanation",
				Plural:    4,
			},
			CollapsibleItems: collapsibleContentItems,
		}
	}

	hasMethodologies := false
	if d.Methodologies != nil {
		for _, meth := range *d.Methodologies {
			p.DatasetLandingPage.Methodologies = append(p.DatasetLandingPage.Methodologies, datasetLandingPageCensus.Methodology{
				Title:       meth.Title,
				URL:         meth.URL,
				Description: meth.Description,
			})
		}
		hasMethodologies = true
	}

	p.Breadcrumb = []coreModel.TaxonomyNode{
		{
			Title: "Home",
			URI:   "/",
		},
		{
			Title: "Census",
			URI:   "/census",
		},
		{
			Title: "Datasets",
			URI:   "/census/datasets",
		},
	}

	sections := make(map[string]coreModel.ContentSection)
	displayOrder := make([]string, 0)

	p.TableOfContents = coreModel.TableOfContents{
		AriaLabel: coreModel.Localisation{
			LocaleKey: "ContentsAria",
			Plural:    1,
		},
		Title: coreModel.Localisation{
			LocaleKey: "Contents",
			Plural:    1,
		},
	}

	sections["summary"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "Summary",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "summary")

	sections["variables"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "Variables",
			Plural:    4,
		},
	}
	displayOrder = append(displayOrder, "variables")

	sections["get-data"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "GetData",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "get-data")

	if len(version.Downloads) > 0 && !isFilterOutput {
		p.DatasetLandingPage.HasDownloads = true
	}

	if isFilterOutput && len(filter.Downloads) > 0 {
		p.DatasetLandingPage.HasDownloads = true
	}

	if p.HasContactDetails {
		sections["contact"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "ContactDetails",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "contact")
	}

	sections["stats-disclosure"] = coreModel.ContentSection{
		Title: coreModel.Localisation{
			LocaleKey: "StatisticalDisclosureControl",
			Plural:    1,
		},
	}
	displayOrder = append(displayOrder, "stats-disclosure")

	if hasMethodologies {
		sections["methodology"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "Methodology",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "methodology")
	}

	if hasOtherVersions {
		sections["version-history"] = coreModel.ContentSection{
			Title: coreModel.Localisation{
				LocaleKey: "VersionHistory",
				Plural:    1,
			},
		}
		displayOrder = append(displayOrder, "version-history")

		for _, ver := range allVersions {
			var version sharedModel.Version
			version.VersionNumber = ver.Version
			version.ReleaseDate = ver.ReleaseDate
			versionUrl := helpers.DatasetVersionUrl(ver.Links.Dataset.ID, ver.Edition, strconv.Itoa(ver.Version))
			version.VersionURL = versionUrl
			version.IsCurrentPage = versionUrl == req.URL.Path
			mapCorrectionAlert(&ver, &version)

			p.Versions = append(p.Versions, version)
		}

		sort.Slice(p.Versions, func(i, j int) bool { return p.Versions[i].VersionNumber > p.Versions[j].VersionNumber })

		if latestVersionNumber != version.Version && hasOtherVersions {
			p.DatasetLandingPage.Panels = append(p.DatasetLandingPage.Panels, datasetLandingPageCensus.Panel{
				IsCorrection: false,
			})
		}
		p.DatasetLandingPage.LatestVersionURL = latestVersionURL
	}

	p.TableOfContents.Sections = sections
	p.TableOfContents.DisplayOrder = displayOrder

	p.DatasetLandingPage.ShareDetails.Language = lang
	currentUrl := helpers.GetCurrentUrl(lang, p.SiteDomain, req.URL.Path)
	p.DatasetLandingPage.DatasetURL = currentUrl

	p.DatasetLandingPage.ShareDetails.ShareLocations = []datasetLandingPageCensus.Share{
		{
			Title: "Facebook",
			Link:  helpers.GenerateSharingLink("facebook", currentUrl, d.Title),
			Icon:  "facebook",
		},
		{
			Title: "Twitter",
			Link:  helpers.GenerateSharingLink("twitter", currentUrl, d.Title),
			Icon:  "twitter",
		},
		{
			Title: "LinkedIn",
			Link:  helpers.GenerateSharingLink("linkedin", currentUrl, d.Title),
			Icon:  "linkedin",
		},
		{
			Title: "Email",
			Link:  helpers.GenerateSharingLink("email", currentUrl, d.Title),
			Icon:  "email",
		},
	}

	p.BetaBannerEnabled = true

	if len(opts) > 0 && !isFilterOutput {
		p.DatasetLandingPage.Dimensions = mapCensusOptionsToDimensions(version.Dimensions, opts, queryStrValues, req.URL.Path, isFlex)
		coverage := []sharedModel.Dimension{
			{
				IsCoverage:        true,
				IsDefaultCoverage: true,
				Title:             Coverage,
				Name:              strings.ToLower(Coverage),
				ShowChange:        isFlex,
				ID:                strings.ToLower(Coverage),
			},
		}
		temp := append(coverage, p.DatasetLandingPage.Dimensions[1:]...)
		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions[:1], temp...)
	}

	if isFilterOutput {
		p.DatasetLandingPage.Dimensions = mapFilterOutputDims(filter, queryStrValues, req.URL.Path)
		coverage := []sharedModel.Dimension{
			{
				IsCoverage:        true,
				IsDefaultCoverage: hasNoAreaOptions,
				Title:             Coverage,
				Name:              strings.ToLower(Coverage),
				ID:                strings.ToLower(Coverage),
				Values:            filter.Dimensions[0].Options,
				ShowChange:        true,
				ChangeURL:         fmt.Sprintf("/filters/%s/dimensions/geography/coverage", filter.FilterID),
			},
		}
		temp := append(coverage, p.DatasetLandingPage.Dimensions[1:]...)
		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions[:1], temp...)
		p.DatasetLandingPage.IsFlexibleForm = false
	}

	if isValidationError {
		p.Error.Title = fmt.Sprintf("Error: %s", d.Title)
	}

	p.BackTo = coreModel.BackTo{
		Text: coreModel.Localisation{
			LocaleKey: "BackToContents",
			Plural:    4,
		},
		AnchorFragment: "toc",
	}

	return p
}

func mapCensusOptionsToDimensions(dims []dataset.VersionDimension, opts []dataset.Options, queryStrValues []string, path string, isFlex bool) []sharedModel.Dimension {
	dimensions := []sharedModel.Dimension{}
	for _, opt := range opts {
		var pDim sharedModel.Dimension

		for _, dimension := range dims {
			if dimension.Name == opt.Items[0].DimensionID {
				pDim.Name = dimension.Name
				pDim.Description = dimension.Description
				pDim.IsAreaType = helpers.IsBoolPtr(dimension.IsAreaType)
				pDim.ShowChange = pDim.IsAreaType && isFlex
				pDim.Title = dimension.Label
				pDim.ID = dimension.ID
			}
		}

		pDim.TotalItems = opt.TotalCount
		midFloor, midCeiling := getTruncationMidRange(opt.TotalCount)

		var displayedOptions []dataset.Option
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) {
			displayedOptions = opt.Items[:3]
			displayedOptions = append(displayedOptions, opt.Items[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, opt.Items[len(opt.Items)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = opt.Items
		}

		for _, opt := range displayedOptions {
			pDim.Values = append(pDim.Values, opt.Label)
		}

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		dimensions = append(dimensions, pDim)
	}
	return dimensions
}

func mapFilterOutputDims(filter filter.Model, queryStrValues []string, path string) []sharedModel.Dimension {
	dimensions := []sharedModel.Dimension{}
	for _, dim := range filter.Dimensions {
		var isAreaType bool
		if helpers.IsBoolPtr(dim.IsAreaType) {
			isAreaType = true
		}
		pDim := sharedModel.Dimension{}
		pDim.Title = dim.Label
		pDim.ID = dim.ID
		pDim.IsAreaType = isAreaType
		pDim.ShowChange = isAreaType
		if isAreaType {
			pDim.ChangeURL = strings.ToLower(fmt.Sprintf("/filters/%s/dimensions/%s", filter.FilterID, dim.Name))
		}
		pDim.TotalItems = len(dim.Options)
		midFloor, midCeiling := getTruncationMidRange(pDim.TotalItems)

		var displayedOptions []string
		if pDim.TotalItems > 9 && !helpers.HasStringInSlice(pDim.ID, queryStrValues) && !pDim.IsAreaType {
			displayedOptions = dim.Options[:3]
			displayedOptions = append(displayedOptions, dim.Options[midFloor:midCeiling]...)
			displayedOptions = append(displayedOptions, dim.Options[len(dim.Options)-3:]...)
			pDim.IsTruncated = true
		} else {
			displayedOptions = dim.Options
		}

		pDim.Values = append(pDim.Values, displayedOptions...)

		q := url.Values{}
		if pDim.IsTruncated {
			q.Add(queryStrKey, pDim.ID)
		}
		pDim.TruncateLink = generateTruncatePath(path, pDim.ID, q)
		dimensions = append(dimensions, pDim)
	}
	return dimensions
}

// getTruncationMidRange returns ints that can be used as the truncation mid range
func getTruncationMidRange(total int) (int, int) {
	mid := total / 2
	midFloor := mid - 2
	midCeiling := midFloor + 3
	if midFloor < 0 {
		midFloor = 0
	}
	return midFloor, midCeiling
}

// generateTruncatePath returns the path to truncate or show all
func generateTruncatePath(path, dimID string, q url.Values) string {
	truncatePath := path
	if q.Encode() != "" {
		truncatePath += fmt.Sprintf("?%s", q.Encode())
	}
	if dimID != "" {
		truncatePath += fmt.Sprintf("#%s", dimID)
	}
	return truncatePath
}
