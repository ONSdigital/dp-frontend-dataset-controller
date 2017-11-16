package mapper

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/ONSdigital/dp-frontend-models/model/datasetEditionsList"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
)

var dimensionTitleMapper = map[string]string{
	"aggregate": "Goods and Services",
	"time":      "Time",
	"geography": "Geographic Areas",
}

// TimeSlice allows sorting of a list of time.Time
type TimeSlice []time.Time

func (p TimeSlice) Len() int {
	return len(p)
}

func (p TimeSlice) Less(i, j int) bool {
	return p[i].Before(p[j])
}

func (p TimeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// CreateFilterableLandingPage creates a filterable dataset landing page based on api model responses
func CreateFilterableLandingPage(d dataset.Model, versions []dataset.Version, datasetID string, opts []dataset.Options) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	if len(d.Contacts) > 0 {
		p.Metadata.Footer.Contact = d.Contacts[0].Name
		p.ContactDetails.Name = d.Contacts[0].Name
		p.ContactDetails.Telephone = d.Contacts[0].Telephone
		p.ContactDetails.Email = d.Contacts[0].Email
	}

	p.Metadata.Footer.DatasetID = datasetID
	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	if len(versions) > 0 {
		p.DatasetLandingPage.DatasetLandingPage.ReleaseDate = versions[0].ReleaseDate
		p.DatasetLandingPage.Edition = versions[0].Edition
		for _, ver := range versions {
			var v datasetLandingPageFilterable.Version
			v.Title = d.Title
			v.Description = d.Description
			v.Edition = ver.Edition
			v.Version = strconv.Itoa(ver.Version)
			v.ReleaseDate = ver.ReleaseDate

			for k, download := range ver.Downloads {
				if len(download.URL) > 0 {
					v.Downloads = append(v.Downloads, datasetLandingPageFilterable.Download{
						Extension: k,
						Size:      download.Size,
						URI:       download.URL,
					})
				}
			}

			p.DatasetLandingPage.Versions = append(p.DatasetLandingPage.Versions, v)
		}
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			if len(opt.Items) < 2 {
				continue
			}

			var pDim datasetLandingPageFilterable.Dimension

			pDim.Title = dimensionTitleMapper[opt.Items[0].DimensionID]
			versionURL, err := url.Parse(d.Links.LatestVersion.URL)
			if err != nil {
				log.Error(err, nil)
			}
			pDim.OptionsURL = fmt.Sprintf("%s/dimensions/%s/options", versionURL.Path, opt.Items[0].DimensionID)

			if opt.Items[0].DimensionID == "time" {
				var ts TimeSlice
				for _, val := range opt.Items {
					t, err := convertMMMYYToTime(val.Label)
					if err != nil {
						log.Error(err, nil)
					}
					ts = append(ts, t)
				}
				sort.Sort(ts)

				startDate := ts[0]

				for i, t := range ts {
					if i != len(ts)-1 {
						if ((ts[i+1].Month() - t.Month()) == 1) || (t.Month() == 12 && ts[i+1].Month() == 1) {
							continue
						}
						pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
						startDate = ts[i+1]
					} else {
						pDim.Values = append(pDim.Values, fmt.Sprintf("All months between %s %d and %s %d", startDate.Month().String(), startDate.Year(), t.Month().String(), t.Year()))
					}
				}

			} else {

				for i, val := range opt.Items {
					if i > 10 {
						break
					}
					pDim.Values = append(pDim.Values, val.Label)
				}

			}

			p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, pDim)
		}
	}

	return p
}

// CreateEditionsList creates a editions list page based on api model responses
func CreateEditionsList(d dataset.Model, editions []dataset.Edition, datasetID string) datasetEditionsList.Page {
	p := datasetEditionsList.Page{}
	p.Type = "dataset_edition_list"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.TaxonomyDomain = os.Getenv("TAXONOMY_DOMAIN")

	if len(d.Contacts) > 0 {
		p.Metadata.Footer.Contact = d.Contacts[0].Name
		p.ContactDetails.Name = d.Contacts[0].Name
		p.ContactDetails.Telephone = d.Contacts[0].Telephone
		p.ContactDetails.Email = d.Contacts[0].Email
	}

	p.Metadata.Footer.DatasetID = datasetID
	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID

	if len(editions) > 0 {
		for _, edition := range editions {

			var latestVersionURL, err = url.Parse(edition.Links.LatestVersion.URL)
			if err != nil {
				log.Error(err, nil)
			}
			var latestVersionPath = latestVersionURL.Path
			fmt.Println(latestVersionPath)

			var e datasetEditionsList.Edition
			e.Title = edition.Edition
			e.LatestVersionURL = latestVersionPath

			p.Editions = append(p.Editions, e)
		}
	}

	return p
}

func convertMMMYYToTime(input string) (t time.Time, err error) {
	return time.Parse("Jan-06", input)
}
