package controller

import (
	"net/http"
	"pr-mail/app/service"
	"pr-mail/pkg/api"
	"pr-mail/pkg/e"
)

type PrController interface {
	Login(w http.ResponseWriter, r *http.Request)
	SaveEmployeePR(w http.ResponseWriter, r *http.Request)
	GeneratePRDetails(w http.ResponseWriter, r *http.Request)
	GeneratePRReport(w http.ResponseWriter, r *http.Request)
	SendPRMail(w http.ResponseWriter, r *http.Request)
}

type PrControllerImpl struct {
	prService service.PrService
}

func NewPrController(prService service.PrService) PrController {
	return &PrControllerImpl{
		prService: prService,
	}
}

func (c *PrControllerImpl) Login(w http.ResponseWriter, r *http.Request) {
	resp, err := c.prService.LoginUser(r)
	if err != nil {
		apiErr := e.NewAPIError(err, "failed to login admin")
		api.Fail(w, apiErr.StatusCode, apiErr.Code, apiErr.Message, err.Error())
		return
	}
	api.Success(w, http.StatusOK, resp)
}

func (c *PrControllerImpl) SaveEmployeePR(w http.ResponseWriter, r *http.Request) {
	err := c.prService.SaveEmployeePR(r)
	if err != nil {
		apiErr := e.NewAPIError(err, "failed to save employee pr")
		api.Fail(w, apiErr.StatusCode, apiErr.Code, apiErr.Message, err.Error())
		return
	}
	api.Success(w, http.StatusOK, "success")
}

func (c *PrControllerImpl) GeneratePRDetails(w http.ResponseWriter, r *http.Request) {
	resp, err := c.prService.GeneratePRDetails(r)
	if err != nil {
		apiErr := e.NewAPIError(err, "failed to generate pr details")
		api.Fail(w, apiErr.StatusCode, apiErr.Code, apiErr.Message, err.Error())
		return
	}
	api.Success(w, http.StatusOK, resp)
}

func (c *PrControllerImpl) GeneratePRReport(w http.ResponseWriter, r *http.Request) {
	resp, err := c.prService.GeneratePRReport(r)
	if err != nil {
		apiErr := e.NewAPIError(err, "failed to generate pr report")
		api.Fail(w, apiErr.StatusCode, apiErr.Code, apiErr.Message, err.Error())
		return
	}
	api.Success(w, http.StatusOK, resp)
}

func (c *PrControllerImpl) SendPRMail(w http.ResponseWriter, r *http.Request) {
	err := c.prService.SendPRMail(r)
	if err != nil {
		apiErr := e.NewAPIError(err, "failed to send mail")
		api.Fail(w, apiErr.StatusCode, apiErr.Code, apiErr.Message, err.Error())
		return
	}
	api.Success(w, http.StatusOK, "success")
}
