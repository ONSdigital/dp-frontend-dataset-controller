package model

//CookiesPolicy contains data for the users cookie policy
type CookiesPolicy struct {
	Essential bool `json:"essential"`
	Usage     bool `json:"usage"`
}
