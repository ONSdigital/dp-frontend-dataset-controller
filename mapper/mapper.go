package mapper

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/data"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageFilterable"
	"github.com/ONSdigital/go-ns/zebedee/zebedeeMapper"
)

// CreateFilterableLandingPage ...
func CreateFilterableLandingPage(ds []data.Dataset, dims []data.Dimension, sp zebedeeMapper.StaticDatasetLandingPage) datasetLandingPageFilterable.Page {
	p := datasetLandingPageFilterable.Page{}
	p.Type = sp.Type
	p.URI = sp.URI
	p.Metadata = sp.Metadata
	p.DatasetLandingPage.DatasetLandingPage = sp.DatasetLandingPage
	p.Breadcrumb = sp.Breadcrumb
	p.ContactDetails = sp.ContactDetails

	for i, d := range ds {
		var v datasetLandingPageFilterable.Version
		v.Title = d.Title
		v.Description = sp.DatasetLandingPage.MetaDescription
		v.Edition = d.Edition
		v.Version = d.Version
		v.ReleaseDate = d.ReleaseDate

		if len(sp.DatasetLandingPage.Datasets)-1 >= i {
			for _, download := range sp.DatasetLandingPage.Datasets[i].Downloads {
				dwnld := datasetLandingPageFilterable.Download(download)
				v.Downloads = append(v.Downloads, dwnld)
			}
		}

		p.DatasetLandingPage.Versions = append(p.DatasetLandingPage.Versions, v)
	}

	for _, dim := range dims {
		var pDim datasetLandingPageFilterable.Dimension

		pDim.Title = dim.Name
		pDim.Values = dim.Values

		p.DatasetLandingPage.Dimensions = append(p.DatasetLandingPage.Dimensions, pDim)
	}

	return p
}
