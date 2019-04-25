package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// SaveToTempFile stores data from a reader to a temporary file
func SaveToTempFile(r io.Reader) (string, error) {
	f, err := ioutil.TempFile("/tmp", "image.*")
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}

	return f.Name(), nil
}

// ImageThumbnail generates a thumbnail given an image
func ImageThumbnail(w io.Writer, format string, r io.Reader) error {
	cmd := exec.Command(
		"convert",
		format+":-",
		"-resize", "100x100",
		"-gravity", "center",
		"-extent", "100x100",
		"png:-",
	)

	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// TextThumbnail generates a thumbnail given text
func TextThumbnail(w io.Writer, text string) error {
	cmd := exec.Command(
		"convert",
		"-size", "60x60",
		"-background", "#eff0f1",
		"-fill", "#101094",
		"-gravity", "center",
		"label:"+text,
		"-extent", "100x100",
		"png:-",
	)

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

var ErrBadRequest = errors.New("bad request")

// Handler handles incomming http requests
func Handler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return ErrBadRequest
	}

	contentType := strings.SplitN(r.Header.Get("Content-Type"), "/", 2)
	if len(contentType) != 2 {
		return ErrBadRequest
	}

	switch contentType[0] {
	case "image":
		if err := ImageThumbnail(w, contentType[1], r.Body); err != nil {
			return err
		}
	default:
		if len(contentType[1]) > 4 {
			return ErrBadRequest
		}

		if err := TextThumbnail(w, contentType[1]); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	http.ListenAndServe(
		":"+os.Getenv("PORT"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			if err := Handler(w, r); err == ErrBadRequest {
				w.Header().Del("Content-Type")
				w.WriteHeader(400)
			} else if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(500)
				fmt.Fprintln(w, err.Error())
			}
		}),
	)
}
