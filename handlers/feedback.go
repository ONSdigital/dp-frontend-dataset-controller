package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/schema"
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
	getFeedback(w, req, false)
}

func getFeedback(w http.ResponseWriter, req *http.Request, hasError bool) {
	var p model.Page

	p.Metadata.Title = "Feedback"

	if hasError {
		p.ServiceMessage = "Description can't be blank"
	}

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
func AddFeedback(auth smtp.Auth, mailAddr, to, from string, isPositive bool) http.HandlerFunc {
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

		if f.Description == "" && !isPositive {
			getFeedback(w, req, true)
			return
		}

		if f.URL == "" {
			f.URL = "Whole site"
		}

		if err := smtp.SendMail(
			mailAddr,
			auth,
			from,
			[]string{to},
			generateFeedbackMessage(f, from, to, isPositive),
		); err != nil {
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Debug("feedback email sent", nil)

		redirectURL := "/feedback/thanks?returnTo=" + f.URL
		http.Redirect(w, req, redirectURL, 301)
	}
}

func generateFeedbackMessage(f Feedback, from, to string, isPositive bool) []byte {
	var description string
	if isPositive {
		description = "Positive feedback received"
	} else {
		description = f.Description
	}

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("From: %s\n", from))
	b.WriteString(fmt.Sprintf("To: %s\n", to))
	b.WriteString(fmt.Sprintf("Subject: Feedback received\n\n"))

	b.WriteString(fmt.Sprintf("Page URL: %s\n", f.URL))
	b.WriteString(fmt.Sprintf("Description: %s\n", description))

	if len(f.Name) > 0 {
		b.WriteString(fmt.Sprintf("Name: %s\n", f.Name))
	}

	if len(f.Email) > 0 {
		b.WriteString(fmt.Sprintf("Email address: %s\n", f.Email))
	}

	return b.Bytes()
}
