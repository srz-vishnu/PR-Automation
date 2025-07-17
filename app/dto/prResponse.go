package dto

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
)

type PRDetailsEmployeeID struct {
	StaffID string `json:"staff_id" validate:"required"`
}

// New: Struct for a single PR's details
// Used for multiple PRs in PRDetailsResponse
type SinglePRDetails struct {
	Owner        string `json:"owner"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	IsMerged     bool   `json:"is_merged"`
	Files        int    `json:"files_changed"`
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
	CommitCount  int    `json:"commit_count"`
	Branch       string `json:"branch"`
	PRLink       string `json:"pr_link"`
}

type PRDetailsResponse struct {
	PRs []SinglePRDetails `json:"prs"`
}

func (args *PRDetailsEmployeeID) Parse(r *http.Request) error {
	strID := chi.URLParam(r, "id")
	if strID == "" {
		return fmt.Errorf("id parameter is missing or empty")
	}
	args.StaffID = strID
	return nil
}

func (args *PRDetailsEmployeeID) Validate() error {
	validate := validator.New()
	err := validate.Struct(args)
	if err != nil {
		return err
	}
	return nil
}
