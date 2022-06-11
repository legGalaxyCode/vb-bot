package utilities

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func ErrorCheck(err error) error {
	if err != nil {
		return fmt.Errorf("error: %s", err.Error())
	}
	return nil
}

func DownloadFile(filepath string, url string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error: %s", err.Error())
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error: %s", err.Error())
	}
	return nil
}
