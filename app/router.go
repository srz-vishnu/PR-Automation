package app

import (
	"pr-mail/app/controller"
	"pr-mail/app/repo"
	"pr-mail/app/service"

	"pr-mail/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func APIRouter(db *gorm.DB) chi.Router {
	r := chi.NewRouter()

	// part
	prRepo := repo.NewPrRepo(db)
	prService := service.NewPrService(prRepo)
	prController := controller.NewPrController(prService)

	//user
	r.Route("/pr", func(r chi.Router) {
		r.Post("/login", prController.Login)
		r.Post("/save", prController.SaveEmployeePR)
		r.Get("/details/{id}", prController.GeneratePRDetails)
		r.Get("/report/{id}", prController.GeneratePRReport)
		r.Post("/mail", prController.SendPRMail)

		//r.With(middleware.JWTAuthMiddleware).Get("/hello", urController.ExampleHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware) // Applying JWT middleware

	})

	return r
}
