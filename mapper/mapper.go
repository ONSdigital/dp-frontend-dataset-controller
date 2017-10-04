package mapper

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"time"

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

func getVersionFromURL(path string) (string, error) {
	lvReg := regexp.MustCompile(`^\/datasets\/.+\/editions\/.+\/versions\/(.+)$`)

	subs := lvReg.FindStringSubmatch(path)
	if len(subs) < 2 {
		return "", errors.New("could not extract version from path")
	}
	return subs[1], nil
}

// CreateFilterableLandingPage ...
func CreateFilterableLandingPage(d dataset.Model, versions []dataset.Version, datasetID string, opts []dataset.Options) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.Metadata.Footer.Contact = d.Contacts[0].Name
	p.Metadata.Footer.DatasetID = datasetID
	p.ContactDetails.Name = d.Contacts[0].Name
	p.ContactDetails.Telephone = d.Contacts[0].Telephone
	p.ContactDetails.Email = d.Contacts[0].Email
	p.DatasetLandingPage.DatasetLandingPage.NextRelease = d.NextRelease
	p.DatasetLandingPage.DatasetID = datasetID
	p.DatasetLandingPage.DatasetLandingPage.ReleaseDate = versions[0].ReleaseDate

	for _, ver := range versions {
		uri, err := url.Parse(ver.Links.Self.URL)
		if err != nil {
			log.Error(err, nil)
		}

		var v datasetLandingPageFilterable.Version
		v.Title = d.Title
		v.Description = d.Description
		v.Edition = ver.Edition
		v.Version, err = getVersionFromURL(uri.Path)
		if err != nil {
			log.Error(err, log.Data{"path": uri.Path})
		}
		v.ReleaseDate = ver.ReleaseDate

		/*if len(sp.DatasetLandingPage.Datasets)-1 >= i {
			for _, download := range sp.DatasetLandingPage.Datasets[i].Downloads {
				dwnld := datasetLandingPageFilterable.Download(download)
				v.Downloads = append(v.Downloads, dwnld)
			}
		} */

		v.Downloads = append(v.Downloads,
			datasetLandingPageFilterable.Download{
				Size:      "438290",
				Extension: "XLSX",
			},
		)

		p.DatasetLandingPage.Versions = append(p.DatasetLandingPage.Versions, v)
	}

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
				if i > 4 {
					break
				}
				pDim.Values = append(pDim.Values, val.Label)
			}

		}

		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, pDim)
	}

	return p
}

func convertMMMYYToTime(input string) (t time.Time, err error) {
	return time.Parse("Jan-06", input)
}
