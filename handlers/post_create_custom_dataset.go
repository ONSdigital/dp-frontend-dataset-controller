package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-net/v2/handlers"
	"github.com/ONSdigital/log.go/v2/log"
)

// PostCreateCustomDataset controls creating a custom dataset using a population type
func PostCreateCustomDataset(fc FilterClient) http.HandlerFunc {
	return handlers.ControllerHandler(func(w http.ResponseWriter, req *http.Request, lang, collectionID, userAccessToken string) {
		postCreateCustomDataset(w, req, fc, lang, collectionID, userAccessToken)
	})
}

func postCreateCustomDataset(w http.ResponseWriter, req *http.Request, fc FilterClient, lang, collectionID, userAccessToken string) {
	ctx := req.Context()

	form, err := parseChangeDimensionForm(req)
	if err != nil {
		http.Redirect(w, req, fmt.Sprintf("/datasets/create?error=true"), http.StatusMovedPermanently)
		return
	}

	filterId, err := fc.CreateCustomFilter(ctx, userAccessToken, "", form.PopulationType)
	if err != nil {
		log.Error(ctx, "failed to create new custom filter", err, log.Data{
			"population-type": form.PopulationType,
		})
		setStatusCode(ctx, w, err)
		return
	}

	http.Redirect(w, req, fmt.Sprintf("/filters/%s/dimensions", filterId), http.StatusMovedPermanently)
}

// createCustomDatasetForm represents form-data for the PostCreateCustomDataset handler.
type postCreateCustomDatasetForm struct {
	PopulationType string
}

// parseChangeDimensionForm parses form data from a http.Request into a changeDimensionForm.
func parseChangeDimensionForm(req *http.Request) (postCreateCustomDatasetForm, error) {
	err := req.ParseForm()
	if err != nil {
		return postCreateCustomDatasetForm{}, fmt.Errorf("error parsing form: %w", err)
	}

	pop := req.FormValue("populationType")
	if pop == "" {
		return postCreateCustomDatasetForm{}, errors.New("missing required value 'populationType'")
	}

	return postCreateCustomDatasetForm{
		PopulationType: pop,
	}, nil
}
