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
		http.Error(w, "Internal Server Error - pars", http.StatusInternalServerError)
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
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}

		var URLs []URLInfo
		err = json.Unmarshal([]byte(string(body)), &URLs)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}

		app.render(w, files, URLs)
	}

	if r.Method == http.MethodPost && r.FormValue("create") == "Create" {
		newLongURL := &URLRequest{
			LongUrl: r.PostFormValue("longURL"),
		}

		endpointUrl := fmt.Sprintf("%s/create", app.backendUrl)

		marshal_struct, _ := json.Marshal(newLongURL)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", endpointUrl, bytes.NewBuffer(marshal_struct))
		res, err := client.Do(req)
		if err != nil {
			app.logger.Println(err.Error())
			fmt.Fprintf(w, "Error: %s", err.Error())
		}
		if res.StatusCode == 200 {
			app.logger.Infof("Successfully created short URL for %s", newLongURL.LongUrl)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}

	// if r.Method == http.MethodDelete && r.FormValue("delete") == "Delete" {
	// 	shortUrl := r.PostFormValue("ShortUrl")
	// 	shortCode := shortUrl[len(shortUrl)-7:] // Extract last 7 characters as short code
	// 	endpointUrl := fmt.Sprintf("%s/%s", app.backendUrl, shortCode)

	// 	client := &http.Client{}
	// 	req, _ := http.NewRequest("DELETE", endpointUrl, nil)
	// 	res, err := client.Do(req)
	// 	if err != nil {
	// 		app.logger.Println(err.Error())
	// 		fmt.Fprintf(w, "Error: %s", err.Error())
	// 	}
	// 	if res.StatusCode == 200 {
	// 		app.logger.Infof("Successfully deleted short URL with code %s", shortCode)
	// 		http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	}
	// }
}
