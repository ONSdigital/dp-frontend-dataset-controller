package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/schema"
	"github.com/nlopes/slack"
)

// Feedback represents a user's feedback
type Feedback struct {
	SpecificPage string `schema:"specific-page"`
	WholeSite    string `schema:"whole-site"`
	URI          string `schema:":uri"`
	URL          string `schema:"url"`
	Description  string `schema:"description"`
	Name         string `schema:"name"`
	Email        string `schema:"email"`
}

// FeedbackThanks loads the Feedback Thank you page
func FeedbackThanks(w http.ResponseWriter, req *http.Request) {
	var p model.Page

	p.Metadata.Title = "Thank you for your feedback"
	returnTo := req.URL.Query().Get("returnTo")

	if returnTo == "Whole site" {
		returnTo = "https://www.ons.gov.uk"
	}
	p.Metadata.Description = returnTo

	cfg := config.Get()

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r := renderer.New(cfg.RendererURL)

	templateHTML, err := r.Do("feedback-thanks", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)
}

// GetFeedback handles the loading of a feedback page
func GetFeedback(w http.ResponseWriter, req *http.Request) {
	var p model.Page

	p.Metadata.Title = "Feedback"

	cfg := config.Get()

	b, err := json.Marshal(p)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r := renderer.New(cfg.RendererURL)

	templateHTML, err := r.Do("feedback", b)
	if err != nil {
		log.ErrorR(req, err, log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)
}

// AddFeedback handles a users feedback request and sends a message to slack
func AddFeedback(api *slack.Client, isPositive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		decoder := schema.NewDecoder()

		var f Feedback
		if err := decoder.Decode(&f, req.Form); err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if f.URL == "" {
			f.URL = "Whole site"
		}

		params := slack.PostMessageParameters{
			Attachments: []slack.Attachment{generateFeedbackMessage(f, isPositive)},
		}

		channelID, timestamp, err := api.PostMessage("feedback", "Feedback received", params)
		if err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debug("message sent to slack", log.Data{"channelID": channelID, "time": timestamp})

		redirectURL := "/feedback/thanks?returnTo=" + f.URL
		http.Redirect(w, req, redirectURL, 301)
	}
}

func generateFeedbackMessage(f Feedback, isPositive bool) slack.Attachment {
	var description string
	if isPositive {
		description = "Positive feedback received"
	} else {
		description = f.Description
	}

	attachment := slack.Attachment{
		Text: "Feedback Received",
		Fields: []slack.AttachmentField{
			{
				Title: "Page URL",
				Value: f.URL,
			},
			{
				Title: "Description",
				Value: description,
			},
		},
	}

	if len(f.Name) > 0 {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Name",
			Value: f.Name,
		})
	}

	if len(f.Email) > 0 {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Email",
			Value: f.Email,
		})
	}

	return attachment
}
