package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-frontend-dataset-controller/cache"
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/helpers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-net/v3/handlers"
	"github.com/ONSdigital/dp-net/v3/handlers/response"
	"github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type legacyLandingPage struct {
	ZebedeeClient   ZebedeeClient
	DatasetClient   APIClientsGoDatasetClient
	FilesAPIClient  FilesAPIClient
	RenderClient    RenderClient
	Language        string
	CollectionID    string
	UserAccessToken string
	CacheList       *cache.List
	Config          config.Config
}

// LegacyLanding will load a zebedee landing page
func LegacyLanding(zc ZebedeeClient, dc APIClientsGoDatasetClient, fc FilesAPIClient, rend RenderClient, cacheList *cache.List, cfg config.Config) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		lp := legacyLandingPage{
			ZebedeeClient:   zc,
			DatasetClient:   dc,
			FilesAPIClient:  fc,
			RenderClient:    rend,
			Language:        lang,
			CollectionID:    collectionID,
			UserAccessToken: userAccessToken,
			CacheList:       cacheList,
			Config:          cfg,
		}
		lp.Build(w, req)
	})
}

func (lp legacyLandingPage) Build(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	ctx := req.Context()

	if handleRequestForZebedeeJSONData(ctx, w, lp.ZebedeeClient, path, lp.UserAccessToken) {
		return
	}

	dlp, err := lp.getDatasetLandingPage(ctx, path)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	bc, err := lp.getBreadcrumb(ctx, dlp.URI)
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	homepageContent, err := lp.getHomepageContent(ctx)
	if err != nil {
		log.Warn(ctx, "unable to get homepage content", log.FormatErrors([]error{err}), log.Data{"homepage_content": err})
	}

	datasets, err := lp.getDatasets(ctx, dlp, log.Data{"path": path})
	if err != nil {
		setStatusCode(ctx, w, err)
		return
	}

	lp.getRelatedDatasetLinks(req.Context(), &dlp)

	// get cached navigation data
	locale := request.GetLocaleCode(req)
	navigationCache, err := lp.CacheList.Navigation.GetNavigationData(ctx, locale)
	if err != nil {
		log.Error(ctx, "failed to get navigation cache", err)
		setStatusCode(ctx, w, err)
		return
	}

	basePage := lp.RenderClient.NewBasePageModel()
	m := mapper.CreateLegacyDatasetLanding(ctx, basePage, req, dlp, bc, datasets, lp.Language, homepageContent.ServiceMessage, homepageContent.EmergencyBanner, navigationCache)

	m.DatasetLandingPage.OSRLogo = helpers.GetOSRLogoDetails(m.Language)

	b, err := json.Marshal(m)
	if err != nil {
		log.Error(ctx, "error marshalling legacy dataset page model", err)
		setStatusCode(ctx, w, err)
		return
	}

	generatedETag := response.GenerateETag(b, true)
	response.SetETag(w, generatedETag)

	lp.RenderClient.BuildPage(w, m, "static-legacy")
}

func (lp legacyLandingPage) getDatasets(ctx context.Context, dlp zebedee.DatasetLandingPage, logData log.Data) ([]zebedee.Dataset, error) {
	datasets := make([]zebedee.Dataset, len(dlp.Datasets))
	errs, ctx := errgroup.WithContext(ctx)
	for i := range dlp.Datasets {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		ld := logData
		ld["dataset_uri"] = dlp.Datasets[i].URI
		errs.Go(func() error {
			d, err := lp.ZebedeeClient.GetDataset(ctx, lp.UserAccessToken, lp.CollectionID, lp.Language, dlp.Datasets[i].URI)
			if err != nil {
				log.Error(ctx, "zebedee client legacy dataset returned an error", err, ld)
				return errors.Wrap(err, "zebedee client legacy dataset returned an error")
			}

			d, err = addFileSizesToDataset(ctx, lp.FilesAPIClient, d, lp.UserAccessToken)
			if err != nil {
				log.Error(ctx, "failed to get file size from files API", err, ld)
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
				d.Downloads[i].Size = "0"
				return d, err
			}

			fileSize := strconv.FormatUint(md.SizeInBytes, 10)
			d.Downloads[i].Size = fileSize
		}
	}

	for i, supplementaryFile := range d.SupplementaryFiles {
		if supplementaryFile.URI != "" {
			md, err := fc.GetFile(ctx, supplementaryFile.URI, authToken)
			if err != nil {
				return d, err
			}

			fileSize := strconv.FormatUint(md.SizeInBytes, 10)
			d.SupplementaryFiles[i].Size = fileSize
		}
	}

	return d, nil
}

func (lp legacyLandingPage) getDatasetLandingPage(ctx context.Context, path string) (zebedee.DatasetLandingPage, error) {
	return lp.ZebedeeClient.GetDatasetLandingPage(ctx, lp.UserAccessToken, lp.CollectionID, lp.Language, path)
}

func (lp legacyLandingPage) getBreadcrumb(ctx context.Context, uri string) ([]zebedee.Breadcrumb, error) {
	return lp.ZebedeeClient.GetBreadcrumb(ctx, lp.UserAccessToken, lp.CollectionID, lp.Language, uri)
}

func (lp legacyLandingPage) getHomepageContent(ctx context.Context) (zebedee.HomepageContent, error) {
	return lp.ZebedeeClient.GetHomepageContent(ctx, lp.UserAccessToken, lp.CollectionID, lp.Language, homepagePath)
}

func (lp legacyLandingPage) getRelatedDatasetLinks(ctx context.Context, dlp *zebedee.DatasetLandingPage) {
	relatedFilterableDatasets := make([]zebedee.Link, len(dlp.RelatedFilterableDatasets))
	var wg sync.WaitGroup

	for i, relatedFilterableDataset := range dlp.RelatedFilterableDatasets {
		wg.Add(1)

		go func(ctx context.Context, i int, dc APIClientsGoDatasetClient, relatedFilterableDataset zebedee.Link) {
			defer wg.Done()

			d, err := dc.GetByPath(ctx, lp.UserAccessToken, "", lp.CollectionID, relatedFilterableDataset.URI)
			if err != nil {
				// log error but continue to map data. any datasets that fail won't get mapped and won't be displayed on frontend
				log.Error(ctx, "error fetching dataset details", err, log.Data{"dataset": relatedFilterableDataset.URI})
				return
			}

			relatedFilterableDatasets[i] = zebedee.Link{Title: d.Title, URI: relatedFilterableDataset.URI}
		}(ctx, i, lp.DatasetClient, relatedFilterableDataset)
	}

	wg.Wait()

	dlp.RelatedFilterableDatasets = relatedFilterableDatasets
}
