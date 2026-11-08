package backend

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

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
