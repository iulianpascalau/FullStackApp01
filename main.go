package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"FullStackApp01/api"
	"FullStackApp01/common"
	"FullStackApp01/storage"
	"github.com/multiversx/mx-chain-logger-go/file"

	"github.com/joho/godotenv"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

const (
	helpTemplate = `NAME:
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
	logsDir           = "logs"
	logsPath          = "log"
	logsLifeSpan      = time.Hour * 24
	logsFileLimitInMB = 1024
)

var (
	log    = logger.GetOrCreate("main")
	cliApp *cli.App

	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
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
	cliApp.Flags = []cli.Flag{
		logLevel,
	}
}

func startApp(c *cli.Context) error {
	logfile, err := prepareLogger(c.GlobalString(logLevel.Name))
	if err != nil {
		return err
	}
	defer func() {
		_ = logfile.Close()
	}()

	log.Info("Starting app", "version", c.App.Version)

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)

	err = logfile.ChangeFileLifeSpan(logsLifeSpan, logsFileLimitInMB)
	log.LogIfError(err)

	// Load .env file
	err = godotenv.Load()
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

	// Create a new ServeMux to avoid global state issues if we expand later
	mux := http.NewServeMux()
	mux.HandleFunc("/register", server.HandleRegister)
	mux.HandleFunc("/login", server.HandleLogin)
	mux.HandleFunc("/change-password", server.HandleChangePassword)
	mux.HandleFunc("/counter", server.HandleCounter)

	srv := &http.Server{
		Addr:    backendInterface,
		Handler: mux,
	}

	// Run server in a goroutine
	go func() {
		log.Info("Starting server", "interface", srv.Addr)
		errListen := srv.ListenAndServe()
		if errListen != nil && !errors.Is(errListen, http.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Info("Server exiting")

	return nil
}

func prepareLogger(logLevel string) (common.LoggerFile, error) {
	err := logger.SetLogLevel(logLevel)
	if err != nil {
		return nil, err
	}

	currentPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current path: %w", err)
	}
	args := file.ArgsFileLogging{
		WorkingDir:      currentPath,
		DefaultLogsPath: logsDir,
		LogFilePrefix:   logsPath,
	}

	fileLogging, err := file.NewFileLogging(args)
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	return fileLogging, nil
}
