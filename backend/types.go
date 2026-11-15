package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
)

const (
	scheduleURL = "https://www.landmarktheatres.com/api/gatsby-source-boxofficeapi/schedule"
	moviesURL   = "https://www.landmarktheatres.com/api/gatsby-source-boxofficeapi/movies"
	referer     = "https://www.landmarktheatres.com/now-playing/"
	host        = "www.landmarktheatres.com"

	// %s represents the string date in format - 2026-10-18
	pageDateURL  = "https://cms-assets.webediamovies.pro/prod/landmarktheatres/%s/public/page-data/now-playing/page-data.json"
	theaterIndex = 4

	theaterURL = "https://cms-assets.webediamovies.pro/prod/landmarktheatres/%s/public/page-data/sq/d/%s.json"

	defaultTheaterId       = "G01CH"
	defaultTheaterName     = "Landmark Theatres Pasadena"
	defaultTheaterTimezone = "America/Los_Angeles"
)

var (
	PosterDir     string
	GLOBAL_CLIENT *http.Client
)

type Landmark struct {
	Theaters []Theater
	Movies   []Movie
}

type Theater struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timeZone"`
}

type TheaterQParam struct {
	ID       string `json:"id"`
	Timezone string `json:"timeZone"`
}

type FavoriteTheaterID struct {
	ID string `json:"id"`
}

type Movie struct {
	ID          string
	Title       string
	Description string
	Actors      []string
	Director    string
	Runtime     string
	Genres      string
	Poster      string
	PosterPath  string
	Rating      string
	Trailer     Trailer
	Showtimes   []Showtime
}

func InitLandmark() *Landmark {
	// TODO: setup proxy client just in case

	// fetch theaters and initialize the Landmark struct
	// movies will be fetched from the frontend since it needs to happen more often
	theaters, err := GetTheaters()
	if err != nil { // theaters would be nil here
		slog.Error("failed to get theaters on initLandmark", "error", err)
		theaters = []Theater{{ID: defaultTheaterId, Name: defaultTheaterName, Timezone: defaultTheaterTimezone}}
	}

	return &Landmark{Theaters: theaters}
}

func (l Landmark) String() string {
	b, err := json.MarshalIndent(l, " ", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

// ~~~~~~~~~~~~~~~~~~~~
// Movie Scraping Responses
// ~~~~~~~~~~~~~~~~~~~~

type ScheduleResponse map[string]TheaterSchedule

type TheaterSchedule struct {
	Schedule       map[string]map[string][]Showtime `json:"schedule"`
	MoviesTags     map[string][]string              `json:"moviesTags"`
	ShowtimesDates []string                         `json:"showtimesDates"`
}

type Showtime struct {
	ID        string       `json:"id"`
	StartsAt  string       `json:"startsAt"`
	IsExpired bool         `json:"isExpired"`
	Data      ShowtimeData `json:"data"`
}

type ShowtimeData struct {
	Ticketing []Ticketing `json:"ticketing"`
}

type Ticketing struct {
	URLs     []string `json:"urls"`
	Type     string   `json:"type"`
	Provider string   `json:"provider"`
}

type MovieResponse []RawMovie

type RawMovie struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"synopsis"`
	Casting     []string   `json:"casting"`
	Direction   []string   `json:"direction"`
	Runtime     int        `json:"runtime"`
	Genres      string     `json:"genres"`
	PosterURL   string     `json:"poster"`
	Rating      string     `json:"certificate"`
	Trailer     Trailer    `json:"trailer"`
	Showtimes   []Showtime `json:"showtimes"`
}

type Trailer struct {
	SD string
	HD string
}

// ~~~~~~~~~~~~~~~~~~~~
// Theater Scraping Responses
// ~~~~~~~~~~~~~~~~~~~~

type PageDataResponse struct {
	StaticHashes []string `json:"staticQueryHashes"`
}

type TheaterDataResponse map[string]TheaterData

type TheaterData struct {
	AllTheater struct {
		Nodes []Theater `json:"nodes"`
	} `json:"allTheater"`
}

// ~~~~~~~~~~~~~~~~~~~~
// Methods
// ~~~~~~~~~~~~~~~~~~~~

func (l *Landmark) GetTheaters() []Theater {
	return l.Theaters
}

func (l *Landmark) GetFavoriteTheaterID() (string, error) {
	return RetrieveFavoriteTheaterID()
}

func (l *Landmark) WriteFavoriteTheaterID(id string) error {
	return StoreFavoriteTheaterID(id)
}

func (l *Landmark) RemoveFavoriteTheaterID() error {
	return DeleteFavoriteTheaterID()
}

func (l *Landmark) GetMovies(currentTheater Theater) ([]Movie, error) {
	sr, err := GetSchedule(currentTheater)
	if err != nil {
		return nil, fmt.Errorf("failed to get landmark schedule: %q", err)
	}

	u, err := url.Parse(moviesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse movies url: %q", err)
	}

	params := u.Query()
	params.Add("castingLimit", "3")
	params.Add("basic", "false")

	theaterSchedule, ok := sr[currentTheater.ID]
	if !ok {
		return nil, fmt.Errorf("schedule missing theater: %s", currentTheater.ID)
	}

	for id := range theaterSchedule.Schedule {
		params.Add("ids", id)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build movies request: %q", err)
	}

	req.Header.Add("referer", referer)
	req.Header.Add("host", host)

	resp, err := GLOBAL_CLIENT.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make movies req: %q", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[ Movies ] non 200 status code: %d | %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %q", err)
	}

	var mr MovieResponse
	if err := json.Unmarshal(bodyBytes, &mr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal movie response: %q", err)
	}

	for i := range mr {
		schedule, ok := theaterSchedule.Schedule[mr[i].ID]
		if !ok {
			continue
		}

		for _, showtimes := range schedule {
			mr[i].Showtimes = append(mr[i].Showtimes, showtimes...)
		}
	}

	if err := StoreRawMovies(l, mr); err != nil {
		return nil, fmt.Errorf("failed to parse raw movie data: %q", err)
	}

	// need to return the movies to be handled by the svelte frontend
	return l.Movies, nil
}
