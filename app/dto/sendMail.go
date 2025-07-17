package dto

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"
)

type SendMailRequest struct {
	StaffID []string `json:"staff_id" validate:"required"`
}

func (args *SendMailRequest) Parse(r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&args)
	if err != nil {
		return err
	}
	return nil
}

func (args *SendMailRequest) Validate() error {
	validate := validator.New()
	err := validate.Struct(args)
	if err != nil {
		return err
	}
	return nil
}
