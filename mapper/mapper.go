package mapper

import (
	"net/url"
	"regexp"

	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/go-ns/clients/dataset"
	"github.com/ONSdigital/go-ns/log"
)

func getVersionFromURL(path string) string {
	lvReg := regexp.MustCompile(`^\/datasets\/.+\/editions\/.+\/versions\/(.+)$`)

	subs := lvReg.FindStringSubmatch(path)
	return subs[1]
}

// CreateFilterableLandingPage ...
func CreateFilterableLandingPage(d dataset.Model, versions []dataset.Version, datasetID string) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	p.Type = "dataset_landing_page"
	p.Metadata.Title = d.Title
	p.URI = d.Links.Self.URL
	p.Metadata.Description = d.Description
	p.Metadata.Footer.Contact = d.Contact.Name
	p.Metadata.Footer.DatasetID = datasetID
	p.ContactDetails.Name = d.Contact.Name
	p.ContactDetails.Telephone = d.Contact.Telephone
	p.ContactDetails.Email = d.Contact.Email
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
		v.Version = getVersionFromURL(uri.Path)
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

	/*for _, dim := range dims {
		var pDim datasetLandingPageFilterable.Dimension

		pDim.Title = dim.Name
		pDim.Values = dim.Values

		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, pDim)
	} */

	return p
}
