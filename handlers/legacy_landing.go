package handlers

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"sync"
)

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, dc DatasetClient, fc FilesAPIClient, rend RenderClient, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		legacyLanding(w, req, zc, dc, fc, rend, collectionID, lang, userAccessToken)
	})
}

func legacyLanding(w http.ResponseWriter, req *http.Request, zc ZebedeeClient, dc DatasetClient, fac FilesAPIClient, rend RenderClient, collectionID, lang, userAccessToken string) {
	path := req.URL.Path
	ctx := req.Context()

	if isRequestForZebedeeJsonData(w, req, zc, path, ctx, userAccessToken) {
		return
	}

	dlp, err := zc.GetDatasetLandingPage(ctx, userAccessToken, collectionID, lang, path)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	bc, err := zc.GetBreadcrumb(ctx, userAccessToken, collectionID, lang, dlp.URI)
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	homepageContent, err := zc.GetHomepageContent(ctx, userAccessToken, collectionID, lang, homepagePath)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	datasets, err := getDatasets(ctx, dlp, zc, fac, userAccessToken, collectionID, lang, log.Data{"request": req})
	if err != nil {
		setStatusCode(req, w, err)
		return
	}

	// Check for filterable datasets and fetch details
	if len(dlp.RelatedFilterableDatasets) > 0 {
		relatedFilterableDatasets := make([]zebedee.Link, len(dlp.RelatedFilterableDatasets))
		var wg sync.WaitGroup

		for i, relatedFilterableDataset := range dlp.RelatedFilterableDatasets {
			wg.Add(1)

			go func(ctx context.Context, i int, dc DatasetClient, relatedFilterableDataset zebedee.Link) {
				defer wg.Done()

				d, err := dc.GetByPath(ctx, userAccessToken, "", collectionID, relatedFilterableDataset.URI)
				if err != nil {
					// log error but continue to map data. any datasets that fail won't get mapped and won't be displayed on frontend
					log.Error(req.Context(), "error fetching dataset details", err, log.Data{
						"dataset": relatedFilterableDataset.URI,
					})
					return
				}

				relatedFilterableDatasets[i] = zebedee.Link{Title: d.Title, URI: relatedFilterableDataset.URI}
			}(req.Context(), i, dc, relatedFilterableDataset)
		}

		wg.Wait()
		dlp.RelatedFilterableDatasets = relatedFilterableDatasets
	}

	basePage := rend.NewBasePageModel()
	m := mapper.CreateLegacyDatasetLanding(basePage, ctx, req, dlp, bc, datasets, lang, homepageContent.ServiceMessage, homepageContent.EmergencyBanner)

	rend.BuildPage(w, m, "static")
}

func getDatasets(ctx context.Context, dlp zebedee.DatasetLandingPage, zc ZebedeeClient, fac FilesAPIClient, userAccessToken, collectionID, lang string, logData log.Data) ([]zebedee.Dataset, error) {
	datasets := make([]zebedee.Dataset, len(dlp.Datasets))
	errs, ctx := errgroup.WithContext(ctx)
	for i := range dlp.Datasets {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		errs.Go(func() error {
			d, err := zc.GetDataset(ctx, userAccessToken, collectionID, lang, dlp.Datasets[i].URI)
			if err != nil {
				log.Error(ctx, "zebedee client legacy dataset returned an error", err, logData)
				return errors.Wrap(err, "zebedee client legacy dataset returned an error")
			}

			d, err = addFileSizesToDataset(ctx, fac, d, userAccessToken)
			if err != nil {
				log.Error(ctx, "failed to get file size from files API", err, logData)
				return errors.Wrap(err, "failed to get file size from files API")
			}

			datasets[i] = d
			return nil
		})
	}

	return datasets, errs.Wait()
}

func addFileSizesToDataset(ctx context.Context, fc FilesAPIClient, d zebedee.Dataset, authToken string) (zebedee.Dataset, error) {
	for i, download := range d.Downloads {
		if download.URI != "" {
			md, err := fc.GetFile(ctx, download.URI, authToken)
			if err != nil {
				return d, err
			}

			fileSize := strconv.Itoa(int(md.SizeInBytes))
			d.Downloads[i].Size = fileSize
		}
	}

	for i, supplementaryFile := range d.SupplementaryFiles {
		if supplementaryFile.URI != "" {
			md, err := fc.GetFile(ctx, supplementaryFile.URI, authToken)
			if err != nil {
				return d, err
			}

			fileSize := strconv.Itoa(int(md.SizeInBytes))
			d.SupplementaryFiles[i].Size = fileSize
		}
	}

	return d, nil
}
