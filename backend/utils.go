package backend

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type OXY_CREDS struct {
	OXY_PASS string
	OXY_HOST string
	OXY_PORT string
	OXY_USER string
}

// TODO: replace the .env with OnePassword
func LoadOxyVars(envFS embed.FS) (OXY_CREDS, error) {
	vars, err := envFS.Open(".env")
	if err != nil {
		return OXY_CREDS{}, fmt.Errorf("failed to open env file: ", err)
	}
	defer vars.Close()

	env, err := godotenv.Parse(vars)
	if err != nil {
		return OXY_CREDS{}, fmt.Errorf("failed to parse env vars: ", err)
	}

	var creds OXY_CREDS

	creds.OXY_PASS = env["OXYLABS_PASSWORD"]
	creds.OXY_HOST = env["OXYLABS_PROXY_HOST"]
	creds.OXY_PORT = env["OXYLABS_PROXY_PORT"]
	creds.OXY_USER = env["OXYLABS_USERNAME"]

	return creds, nil
}

func BuildProxyClient(envFS embed.FS) (*http.Client, error) {
	creds, err := LoadOxyVars(envFS)
	if err != nil {
		return nil, err
	}

	// build the proxyURL using the env variables
	proxyString := fmt.Sprintf("http://user-%s:%s@%s:%s", creds.OXY_USER, creds.OXY_PASS, creds.OXY_HOST, creds.OXY_PORT)
	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		return nil, fmt.Errorf("parsing proxy url: %v", err)
	}

	// create transport using the proxyURL
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// wrap the proxy transport in the client
	client := &http.Client{Transport: transport}

	return client, nil
}

func GetPageData() (PageDataResponse, error) {
	today := time.Now().Format("2006-01-02")
	pageDataURL := fmt.Sprintf(pageDateURL, today)

	slog.Info("Page Data URL Built", "URL", pageDataURL)

	req, err := http.NewRequest(http.MethodGet, pageDataURL, nil)
	if err != nil {
		return PageDataResponse{}, fmt.Errorf("failed to create page data request: %q", err)
	}

	resp, err := GLOBAL_CLIENT.Do(req)
	if err != nil {
		return PageDataResponse{}, fmt.Errorf("failed to make page data request: %q", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PageDataResponse{}, fmt.Errorf("non 200 status code for page data response - Code: %d | Status: %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return PageDataResponse{}, fmt.Errorf("failed to read page data response body: %q", err)
	}

	var pageData PageDataResponse

	if err := json.Unmarshal(bodyBytes, &pageData); err != nil {
		return PageDataResponse{}, fmt.Errorf("failed to unmarshal page data response: %q", err)
	}

	if len(pageData.StaticHashes) == 0 {
		return PageDataResponse{}, errors.New("page data response is empty")
	}

	return pageData, nil
}

func GetTheaters() ([]Theater, error) {
	pd, err := GetPageData()
	if err != nil {
		return nil, fmt.Errorf("failed to create page data for theaters: %q", err)
	}

	if len(pd.StaticHashes) <= theaterIndex {
		return nil, errors.New("theater index out of range of static hashes")
	}

	theaterHash := pd.StaticHashes[theaterIndex]
	if theaterHash == "" {
		return nil, errors.New("theater hash is empty")
	}

	today := time.Now().Format("2006-01-02")
	theaterURL := fmt.Sprintf(theaterURL, today, theaterHash)

	req, err := http.NewRequest(http.MethodGet, theaterURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create theater data request: %q", err)
	}

	resp, err := GLOBAL_CLIENT.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make theater data request: %q", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 status code for theater data response - Code: %d | Status: %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read theater data response body: %q", err)
	}

	var theaterDataResponse TheaterDataResponse

	if err := json.Unmarshal(bodyBytes, &theaterDataResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal thater data response: %q", err)
	}

	var theaters []Theater
	for _, theaterData := range theaterDataResponse {
		theaters = append(theaters, theaterData.AllTheater.Nodes...)
	}

	if len(theaters) == 0 {
		return nil, errors.New("theater data response is empty")
	}

	return theaters, nil
}

func GetSchedule(currentTheater Theater) (ScheduleResponse, error) {
	u, err := url.Parse(scheduleURL)
	if err != nil {
		return ScheduleResponse{}, fmt.Errorf("failed to parse url: %q", err)
	}

	// can't use name field in the qparam to collect the schedule
	currentQPTheater := TheaterQParam{ID: currentTheater.ID, Timezone: currentTheater.Timezone}

	theater, err := json.Marshal(currentQPTheater)
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

	resp, err := GLOBAL_CLIENT.Do(req)
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

func StorePosterImage(u string, title string) (string, error) {
	if u == "" {
		return "", errors.New("no poster url found")
	}

	if title == "" {
		return "", errors.New("no title found")
	}

	u2, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	uPath := path.Base(u2.Path)
	imgExt := filepath.Ext(uPath)

	if imgExt != ".jpg" && imgExt != ".jpeg" && imgExt != ".png" {
		return "", fmt.Errorf("non-valid image type for poster [%s]", imgExt)
	}

	cleanTitle := strings.NewReplacer(" ", "_", "/", "-", "\\", "-", ":", "-").Replace(title)
	fileName := cleanTitle + imgExt

	imgPath := filepath.Join(PosterDir, fileName)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := GLOBAL_CLIENT.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(imgPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return imgPath, nil
}

func StoreFavoriteTheaterID(id string) error {
	if id == "" {
		return errors.New("no theater id found")
	}

	fileName := "favorite-theater.json"
	ftPath := filepath.Join(PosterDir, fileName)

	ft := FavoriteTheaterID{ID: id}
	ftBytes, err := json.MarshalIndent(ft, "", "  ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(ftPath, ftBytes, 0o644); err != nil {
		return err
	}

	return nil
}

func DeleteFavoriteTheaterID() error {
	fileName := "favorite-theater.json"
	ftPath := filepath.Join(PosterDir, fileName)

	ft := FavoriteTheaterID{ID: ""}
	ftBytes, err := json.MarshalIndent(ft, "", "  ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(ftPath, ftBytes, 0o644); err != nil {
		return err
	}

	return nil
}

func RetrieveFavoriteTheaterID() (string, error) {
	fileName := "favorite-theater.json"
	ftPath := filepath.Join(PosterDir, fileName)

	ftBytes, err := os.ReadFile(ftPath)
	if err != nil {
		return "", err
	}

	var ft FavoriteTheaterID
	if err := json.Unmarshal(ftBytes, &ft); err != nil {
		return "", err
	}

	return ft.ID, nil
}

func CleanRuntime(r int) string {
	minutes := r / 60
	hours := minutes / 60
	rMinutes := minutes % 60

	return fmt.Sprintf("%d HR, %d MIN", hours, rMinutes)
}

func StoreRawMovies(l *Landmark, mr MovieResponse) error {
	var err error

	// need to make the temp poster directory if one does not exist
	if PosterDir == "" {
		PosterDir, err = os.MkdirTemp("", "Landmark-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary app dir", err)
		}
	}

	// always reset movies on func call
	movies := make([]Movie, 0, len(mr))

	for _, m := range mr {
		movie := Movie{ID: m.ID, Title: m.Title, Description: m.Description, Actors: m.Casting, Director: m.Direction[0], Genres: m.Genres, Rating: m.Rating, Trailer: m.Trailer, Showtimes: m.Showtimes}

		// clean runtime (seconds -> xHR xMIN)
		movie.Runtime = CleanRuntime(m.Runtime)

		// store poster image in tmp dir
		movie.PosterPath, err = StorePosterImage(m.PosterURL, m.Title)
		if err != nil {
			err = fmt.Errorf("failed to store poster url for %s: %q", m.Title, err) // we should still display the movie if we cant collect the poster
		} else {
			// frontend request path
			movie.Poster = "/posters/" + filepath.Base(movie.PosterPath)
		}

		// add movie to temp array
		movies = append(movies, movie)
	}

	// set fresh movies
	l.Movies = movies

	return err // want the frontend to be aware of the non-fatal error
}
