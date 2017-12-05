package feedback

import "github.com/ONSdigital/dp-frontend-models/model"

// Page contains data reused for feedback model
type Page struct {
	model.Page
	Radio       string `json:"radio"`
	Purpose     string `json:"purpose"`
	Feedback    string `json:"feedback"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	ErrorType   string `json:"error_type"`
	PreviousURL string `json:"previous_url"`
}
