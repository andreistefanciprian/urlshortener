package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

type URLRequest struct {
	LongUrl   string `json:"longUrl"`
	ExpiresIn int    `json:"expiresIn,omitempty"`
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
		http.Error(w, "Internal Server Error - template execution failed", http.StatusInternalServerError)
	}
}

func (app *FrontendApp) home(w http.ResponseWriter, r *http.Request) {
	app.logger.Debugf("Request: %s %s", r.Method, r.URL.Path)

	files := []string{
		"./templates/base.tmpl",
		"./templates/pages/home.tmpl",
		"./templates/partials/nav.tmpl",
		"./templates/partials/footer.tmpl",
	}

	if r.Method == http.MethodGet {
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/", app.backendUrl), nil)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		var URLs []URLInfo
		if err := json.Unmarshal(body, &URLs); err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		app.render(w, files, URLs)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("create") == "Create" {
		expiresIn := 7
		if s := r.PostFormValue("expiresIn"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				expiresIn = n
			}
		}
		payload := &URLRequest{
			LongUrl:   r.PostFormValue("longURL"),
			ExpiresIn: expiresIn,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		client := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/create", app.backendUrl), bytes.NewBuffer(body))
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			app.logger.Infof("Created short URL for %s", payload.LongUrl)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		errBody, _ := io.ReadAll(res.Body)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(res.StatusCode)
		_, _ = w.Write(errBody)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("_method") == "DELETE" && r.FormValue("delete") == "Delete" {
		shortUrl := r.PostFormValue("ShortUrl")
		shortCode := shortUrl[len(shortUrl)-7:]
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", app.backendUrl, shortCode), nil)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			app.logger.Infof("Deleted short URL with code %s", shortCode)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func (app *FrontendApp) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *FrontendApp) ready(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(app.backendUrl + "/health")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
