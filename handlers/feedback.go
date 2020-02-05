package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"regexp"

	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/feedback"
	"github.com/ONSdigital/log.go/log"
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
func FeedbackThanks(rendererURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var p model.Page
		ctx := req.Context()

		p.Metadata.Title = "Thank you"
		returnTo := req.URL.Query().Get("returnTo")

		if returnTo == "Whole site" {
			returnTo = "https://www.ons.gov.uk"
		}
		p.Metadata.Description = returnTo

		b, err := json.Marshal(p)
		if err != nil {
			log.Event(ctx, "unable to marshal page data", log.Error(err), log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r := renderer.New(rendererURL)

		templateHTML, err := r.Do("feedback-thanks", b)
		if err != nil {
			log.Event(ctx, "failed to render feedback-thanks template", log.Error(err), log.Data{"setting-response-status": http.StatusInternalServerError})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(templateHTML)
	}
}

// GetFeedback handles the loading of a feedback page
func GetFeedback(rendererURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		getFeedback(w, req, req.Referer(), rendererURL, "", "", "", "", "")
	}
}

func getFeedback(w http.ResponseWriter, req *http.Request, url, rendererURL, errorType, purpose, description, name, email string) {
	var p feedback.Page

	var services = make(map[string]string)
	services["cmd"] = "Customising data by applying filters"
	services["dev"] = "ONS developer website"

	service := services[req.URL.Query().Get("service")]
	if service == "" {
		io.Copy(ioutil.Discard, req.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.ServiceDescription = service

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

	b, err := json.Marshal(p)
	if err != nil {
		log.Event(req.Context(), "unable to marshal feedback page data", log.Error(err), log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r := renderer.New(rendererURL)

	templateHTML, err := r.Do("feedback", b)
	if err != nil {
		log.Event(req.Context(), "failed to render feedback template", log.Error(err), log.Data{"setting-response-status": http.StatusInternalServerError})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(templateHTML)
}

// AddFeedback handles a users feedback request and sends a message to slack
func AddFeedback(auth smtp.Auth, mailAddr, to, from, rendererURL string, isPositive bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if err := req.ParseForm(); err != nil {
			log.Event(ctx, "unable to parse request form", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		decoder := schema.NewDecoder()

		var f Feedback
		if err := decoder.Decode(&f, req.Form); err != nil {
			log.Event(ctx, "unable to decode request form", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if f.FeedbackFormType == "page" && f.Purpose == "" && !isPositive {
			getFeedback(w, req, f.URL, rendererURL, "purpose", f.Purpose, f.Description, f.Name, f.Email)
			return
		}

		if f.Description == "" && !isPositive {
			getFeedback(w, req, f.URL, rendererURL, "description", f.Purpose, f.Description, f.Name, f.Email)
			return
		}

		if len(f.Email) > 0 && !isPositive {
			if ok, err := regexp.MatchString(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,6}$`, f.Email); !ok || err != nil {
				getFeedback(w, req, f.URL, rendererURL, "email", f.Purpose, f.Description, f.Name, f.Email)
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
			log.Event(ctx, "failed to send message", log.Error(err))
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
