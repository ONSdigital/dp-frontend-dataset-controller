package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"regexp"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/mapper"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/feedback"
	"github.com/ONSdigital/go-ns/clients/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/schema"
)

// Feedback represents a user's feedback
type Feedback struct {
	Purpose          string `schema:"purpose"`
	Type             string `schema:"type"`
	URI              string `schema:":uri"`
	URL              string `schema:"url"`
	Description      string `schema:"description"`
	Name             string `schema:"name"`
	Email            string `schema:"email"`
	FeedbackFormType string `schema:"feedback-form-type"`
}

// FeedbackThanks loads the Feedback Thank you page
func FeedbackThanks(w http.ResponseWriter, req *http.Request) {
	var p model.Page
	mapper.SetTaxonomyDomain(&p)

	p.Metadata.Title = "Thank you"
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
	getFeedback(w, req, req.Referer(), "", "", "", "", "")
}

func getFeedback(w http.ResponseWriter, req *http.Request, url, errorType, purpose, description, name, email string) {
	var p feedback.Page
	mapper.SetTaxonomyDomain(&p.Page)

	p.Metadata.Title = "Feedback"
	p.Metadata.Description = url

	if len(p.Metadata.Description) > 50 {
		p.Metadata.Description = p.Metadata.Description[len(p.Metadata.Description)-50 : len(p.Metadata.Description)]
	}

	p.ErrorType = errorType
	p.Purpose = purpose
	p.Feedback = description
	p.Name = name
	p.Email = email
	p.PreviousURL = url

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

		if f.FeedbackFormType == "page" && f.Purpose == "" && !isPositive {
			getFeedback(w, req, f.URL, "purpose", f.Purpose, f.Description, f.Name, f.Email)
			return
		}

		if f.Description == "" && !isPositive {
			getFeedback(w, req, f.URL, "description", f.Purpose, f.Description, f.Name, f.Email)
			return
		}

		if len(f.Email) > 0 && !isPositive {
			if ok, err := regexp.MatchString(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}$`, f.Email); !ok || err != nil {
				getFeedback(w, req, f.URL, "email", f.Purpose, f.Description, f.Name, f.Email)
				return
			}
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

	if len(f.Type) > 0 {
		b.WriteString(fmt.Sprintf("Feedback Type: %s\n", f.Type))
	}

	b.WriteString(fmt.Sprintf("Page URL: %s\n", f.URL))
	b.WriteString(fmt.Sprintf("Description: %s\n", description))
	if len(f.Purpose) > 0 {
		b.WriteString(fmt.Sprintf("Purpose: %s\n", f.Purpose))
	}

	if len(f.Name) > 0 {
		b.WriteString(fmt.Sprintf("Name: %s\n", f.Name))
	}

	if len(f.Email) > 0 {
		b.WriteString(fmt.Sprintf("Email address: %s\n", f.Email))
	}

	return b.Bytes()
}
