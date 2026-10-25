package backend

import (
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
)

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

	cleanTitle := strings.ReplaceAll(title, " ", "_")
	fileName := cleanTitle + imgExt

	imgPath := filepath.Join(PosterDir, fileName)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
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
