package dto

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
)

type PRReportRequest struct {
	StaffID string `json:"staff_id" validate:"required"`
}

type PRReportResponse struct {
	EmpID      string `json:"emp_id"`
	PRLink     string `json:"pr_link"`
	ReportText string `json:"report_text"`
}

func (args *PRReportRequest) Parse(r *http.Request) error {
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

func (args *PRReportRequest) Validate() error {
	validate := validator.New()
	err := validate.Struct(args)
	if err != nil {
		return err
	}
	return nil
}
