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

type PRDetailsResponse struct {
	Owner        string `json:"owner"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"` // open, merged, etc.
	IsMerged     bool   `json:"is_merged"`
	Files        int    `json:"files_changed"`
	LinesAdded   int    `json:"lines_added"`
	LinesRemoved int    `json:"lines_removed"`
	CommitCount  int    `json:"commit_count"`
	Branch       string `json:"branch"`
}

func (args *PRDetailsEmployeeID) Parse(r *http.Request) error {
	strID := chi.URLParam(r, "id")
	if strID == "" {
		return fmt.Errorf("id parameter is missing or empty")
	}
	// intID, err := strconv.Atoi(strID)
	// if err != nil {
	// 	return err
	// }
	// args.StaffID = int64(intID)
	// return nil

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
