package dto

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"
)

type SaveEmployeePRRequest struct {
	PRs []SinglePR `json:"prs" validate:"required,dive"`
}

type SinglePR struct {
	Status  string `json:"status" validate:"required"`
	StaffID string `json:"staff_id" validate:"required"`
	PRLink  string `json:"pr_link" validate:"required,url"`
}

func (args *SaveEmployeePRRequest) Parse(r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&args)
	if err != nil {
		return err
	}
	return nil
}

func (args *SaveEmployeePRRequest) Validate() error {
	validate := validator.New()
	err := validate.Struct(args)
	if err != nil {
		return err
	}
	return nil
}
