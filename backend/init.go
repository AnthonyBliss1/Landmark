package backend

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	PosterDir string // using this in handlers package

	scheduleURL string
	moviesURL   string
	referer     string
	host        string
	theaterId   string
	timezone    string
	client      = &http.Client{}
)

func init() {
	godotenv.Load()

	scheduleURL = os.Getenv("SCHEDULE_URL")
	if scheduleURL == "" {
		log.Fatal("failed to load schedule url")
	}

	moviesURL = os.Getenv("MOVIES_URL")
	if moviesURL == "" {
		log.Fatal("failed to load movies url")
	}

	referer = os.Getenv("REFERER")
	if referer == "" {
		log.Fatal("failed to load referer")
	}

	host = os.Getenv("HOST")
	if host == "" {
		log.Fatal("failed to load host")
	}

	theaterId = os.Getenv("THEATER_ID")
	if theaterId == "" {
		log.Fatal("failed to load theater id")
	}

	timezone = os.Getenv("TIMEZONE")
	if timezone == "" {
		log.Fatal("failed to load timezone")
	}
}
