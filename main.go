package main

import (
	"embed"
	"log"
	"log/slog"
	"net/http"
	"time"

	"changeme/backend"
	"changeme/middleware"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed .env
var embeddedEnv embed.FS

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[bool]("refresh")
	application.RegisterEvent[int]("bookingEventStart")

	var err error

	backend.GLOBAL_CLIENT, err = backend.BuildProxyClient(embeddedEnv)
	if err != nil {
		// fall back to normal client, slog the error
		slog.Error("failed to build proxy client, falling back to regular client: ", err)
		backend.GLOBAL_CLIENT = &http.Client{}
	} else {
		slog.Info("sucessfully built proxy client")
	}
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Landmark",
		Description: "A simple GUI providing movie and show times for Landmark theaters",
		Services: []application.Service{
			application.NewService(backend.InitLandmark()),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: middleware.PosterMiddleWare,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Landmark ARB",
		Frameless: true,
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	// getMovies refresh go routine
	go func() {
		for {
			time.Sleep(time.Hour) // will start with hour refresh maybe bump to 2 hours
			app.Event.Emit("refresh", true)
		}
	}()

	app.Event.On("bookingEventStart", func(e *application.CustomEvent) {
		go func() {
			timerCount := e.Data.(int)
			app.Logger.Info("[ Booking timer set ]", "minutes", timerCount)
			time.Sleep(time.Minute * time.Duration(timerCount))
			window.SetURL("/")
		}()
	})

	// Run the application. This blocks until the application has been exited.
	err := app.Run()
	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
