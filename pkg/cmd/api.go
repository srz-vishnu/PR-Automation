package cmd

import (
	"pr-mail/app"
	gormdb "pr-mail/app/gorm_db"
	"pr-mail/pkg/api"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "Root short description",
	Long:  "Root long description",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Root short description",
	Long:  "Root long description",
	Run:   StartAPI,
}

func StartAPI(*cobra.Command, []string) {
	db, err := gormdb.ConnectDb()
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	r := app.APIRouter(db)
	api.Start(r)

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
