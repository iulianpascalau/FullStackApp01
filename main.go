package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"FullStackApp01/api"
	"FullStackApp01/storage"

	"github.com/joho/godotenv"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

const helpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`

var (
	log    = logger.GetOrCreate("main")
	cliApp *cli.App
)

func main() {
	initCliFlags()

	cliApp.Action = func(c *cli.Context) error {
		return startApp(c)
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func initCliFlags() {
	cliApp = cli.NewApp()
	cli.AppHelpTemplate = helpTemplate
	cliApp.Name = "App backend"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.1", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "App backend"
	cliApp.Authors = []cli.Author{
		{
			Name:  "Iulian Pascalau",
			Email: "iulian.pascalau@gmail.com",
		},
	}
}

func startApp(c *cli.Context) error {
	log.Info("Starting app", "version", c.App.Version)

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	jwtKey := os.Getenv("JWT_KEY")
	if len(jwtKey) == 0 {
		return errors.New("JWT_KEY is not set in the .env file")
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if len(adminPassword) == 0 {
		return errors.New("ADMIN_PASSWORD is not set in the .env file")
	}

	backendInterface := os.Getenv("BACKEND_INTERFACE")
	if len(backendInterface) == 0 {
		return errors.New("BACKEND_INTERFACE is not set in the .env file")
	}

	// Create or open a database in the "data" folder
	store, err := storage.NewStore("data")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()

	// Ensure an admin exists
	_ = store.SaveUser("admin", adminPassword, "admin")

	server := api.NewServer(store, []byte(jwtKey))

	http.HandleFunc("/register", server.HandleRegister)
	http.HandleFunc("/login", server.HandleLogin)
	http.HandleFunc("/change-password", server.HandleChangePassword)
	http.HandleFunc("/counter", server.HandleCounter)

	log.Info("Starting server ", "interface", backendInterface)
	err = http.ListenAndServe(backendInterface, nil)
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}

	return nil
}
