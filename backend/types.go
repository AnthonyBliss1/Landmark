package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Landmark struct {
	Theater
	Movies []Movie
}

type Theater struct {
	ID       string `json:"id"`
	Timezone string `json:"timeZone"`
}

type Movie struct {
	ID          string
	Title       string
	Description string
	Actors      []string
	Director    string
	Runtime     string
	Genres      string
	PosterPath  string
	Rating      string
	Trailer     Trailer
}

func InitLandmark() *Landmark {
	// setup proxy client just in case

	return &Landmark{Theater: Theater{ID: theaterId, Timezone: timezone}}
}

func (l Landmark) String() string {
	b, err := json.MarshalIndent(l, " ", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

// ~~~~~~~~~~~~~~~~~~~~
// Scraping Responses
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
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"synopsis"`
	Casting     []string `json:"casting"`
	Direction   []string `json:"direction"`
	Runtime     int      `json:"runtime"`
	Genres      string   `json:"genres"`
	PosterURL   string   `json:"poster"`
	Rating      string   `json:"certificate"`
	Trailer     Trailer  `json:"trailer"`
}

type Trailer struct {
	SD string
	HD string
}

// ~~~~~~~~~~~~~~~~~~~~
// Methods
// ~~~~~~~~~~~~~~~~~~~~

func (l *Landmark) GetSchedule() (ScheduleResponse, error) {
	u, err := url.Parse(scheduleURL)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to parse url: %q", err)
	}

	theater, err := json.Marshal(l.Theater)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to marhsal theater data: %q", err)
	}

	// Dynamic to and from dates for api params
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	from := fmt.Sprintf("%sT03:00:00", today)
	to := fmt.Sprintf("%sT03:00:00", tomorrow)

	params := u.Query()
	params.Add("from", from)
	params.Add("theaters", string(theater))
	params.Add("to", to)
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to create req: %q", err)
	}

	req.Header.Add("referer", referer)
	req.Header.Add("host", host)

	resp, err := client.Do(req)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to make schedules req: %q", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ScheduleResponse{}, fmt.Errorf("[ Schedule ] non 200 status code: %d | %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to read response body: %q", err)
	}

	var sr ScheduleResponse
	if err := json.Unmarshal(bodyBytes, &sr); err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to unmarshal schedule response: %q", err)
	}

	return sr, nil
}

func (l *Landmark) GetMovies() ([]Movie, error) {
	sr, err := l.GetSchedule()
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

	theaterSchedule, ok := sr[l.ID]
	if !ok {
		return nil, fmt.Errorf("schedule missing theater: %s", l.ID)
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

	resp, err := client.Do(req)
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

	if err := l.StoreRawMovies(mr); err != nil {
		return nil, fmt.Errorf("failed to parse raw movie data: %q", err)
	}

	// need to return the movies to be handled by the svelte frontend
	return l.Movies, nil
}

func (l *Landmark) StoreRawMovies(mr MovieResponse) error {
	var err error

	// need to make the temp poster directory if one does not exist
	if posterDir == "" {
		posterDir, err = os.MkdirTemp("", "Landmark-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary app dir", err)
		}
	}

	for _, m := range mr {
		movie := Movie{ID: m.ID, Title: m.Title, Description: m.Description, Actors: m.Casting, Director: m.Direction[0], Genres: m.Genres, Rating: m.Rating, Trailer: m.Trailer}

		// clean runtime (seconds -> xHR xMIN)
		movie.Runtime = CleanRuntime(m.Runtime)

		// store poster image in tmp dir
		movie.PosterPath, err = StorePosterImage(m.PosterURL, m.Title)
		if err != nil {
			return fmt.Errorf("failed to store poster url for %s: %q", m.Title, err)
		}

		// add movie to appstate
		l.Movies = append(l.Movies, movie)
	}

	return nil
}
