package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FrontendHandler interface {
	// Define methods for handling HTTP requests
	home(w http.ResponseWriter, r *http.Request)
}

type FrontendApp struct {
	logger     *logrus.Logger
	backendUrl string
}

func NewFrontendApp(logger *logrus.Logger, backendUrl string) *FrontendApp {
	return &FrontendApp{
		logger:     logger,
		backendUrl: backendUrl,
	}
}

type URLInfo struct {
	ShortUrl   string                 `json:"shortUrl"`
	LongUrl    string                 `json:"longUrl"`
	CreatedAt  *timestamppb.Timestamp `json:"createdAt"`
	Expiration *timestamppb.Timestamp `json:"expiration,omitempty"`
}

func (app *FrontendApp) render(w http.ResponseWriter, files []string, data interface{}) {
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.logger.Println(err.Error())
		http.Error(w, "Internal Server Error - template parsing failed", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.logger.Println(err.Error())
		http.Error(w, "Internal Server Error - exec templ", http.StatusInternalServerError)
	}
}

type URLRequest struct {
	LongUrl    string                 `json:"longUrl"`
	Expiration *timestamppb.Timestamp `json:"expiration,omitempty"`
}

func (app *FrontendApp) home(w http.ResponseWriter, r *http.Request) {
	// Log all requests to debug multiple calls
	app.logger.Debugf("DEBUG: Request - Method: %s, Path: %s, User-Agent: %s",
		r.Method, r.URL.Path, r.Header.Get("User-Agent"))

	// html go templated files
	files := []string{
		"./templates/base.tmpl",
		"./templates/pages/home.tmpl",
		"./templates/partials/nav.tmpl",
		"./templates/partials/footer.tmpl",
	}

	if r.Method == http.MethodGet {
		client := &http.Client{}
		endpointUrl := fmt.Sprintf("%s/", app.backendUrl)
		req, _ := http.NewRequest("GET", endpointUrl, nil)
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		var URLs []URLInfo
		err = json.Unmarshal(body, &URLs)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}

		app.render(w, files, URLs)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("create") == "Create" {
		newLongURL := &URLRequest{
			LongUrl: r.PostFormValue("longURL"),
		}

		endpointUrl := fmt.Sprintf("%s/create", app.backendUrl)

		marshal_struct, _ := json.Marshal(newLongURL)
		client := &http.Client{}
		app.logger.Debugf("Making POST request to: %s with payload: %s", endpointUrl, string(marshal_struct))
		req, _ := http.NewRequest("POST", endpointUrl, bytes.NewBuffer(marshal_struct))
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			app.logger.Println(err.Error())
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		if res.StatusCode == 200 {
			app.logger.Infof("Successfully created short URL for %s", newLongURL.LongUrl)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Handle DELETE requests (POST with _method=DELETE)
	if r.Method == http.MethodPost && r.FormValue("_method") == "DELETE" && r.FormValue("delete") == "Delete" {
		shortUrl := r.PostFormValue("ShortUrl")
		app.logger.Debugf("Processing delete request for URL: %s", shortUrl)

		shortCode := shortUrl[len(shortUrl)-7:] // Extract last 7 characters as short code
		endpointUrl := fmt.Sprintf("%s/%s", app.backendUrl, shortCode)
		app.logger.Debugf("Making DELETE request to: %s", endpointUrl)

		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpointUrl, nil)
		res, err := client.Do(req)
		if err != nil {
			app.logger.Debugf("Error making delete request: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		defer res.Body.Close()

		app.logger.Debugf("Delete response status: %d", res.StatusCode)
		if res.StatusCode == 200 {
			app.logger.Infof("Successfully deleted short URL with code %s", shortCode)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
