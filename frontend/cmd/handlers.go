package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

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

// stripHTTPScheme removes http:// or https:// from any URL string for display purposes
func stripHTTPScheme(url string) string {
	if len(url) >= 8 && url[:8] == "https://" {
		return url[8:]
	} else if len(url) >= 7 && url[:7] == "http://" {
		return url[7:]
	}
	return url
}

// StripHTTPSchemeFromShortUrl returns the ShortUrl without http:// or https:// for display purposes
func (u *URLInfo) StripHTTPSchemeFromShortUrl() string {
	return stripHTTPScheme(u.ShortUrl)
}

// StripHTTPSchemeFromLongUrl returns the LongUrl without http:// or https:// for display purposes
func (u *URLInfo) StripHTTPSchemeFromLongUrl() string {
	return stripHTTPScheme(u.LongUrl)
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
		req, err := http.NewRequest("GET", endpointUrl, nil)
		if err != nil {
			app.logger.Println("Failed to create HTTP request:", err.Error())
			http.Error(w, "Internal Server Error - failed to create request", http.StatusInternalServerError)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			app.logger.Errorf("Failed to read response body: %v", err)
			return
		}

		var URLs []URLInfo
		err = json.Unmarshal(body, &URLs)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}

		// Render the template with URLs data
		app.render(w, files, URLs)
		return
	}

	if r.Method == http.MethodPost && r.FormValue("create") == "Create" {
		newLongURL := &URLRequest{
			LongUrl: r.PostFormValue("longURL"),
		}

		endpointUrl := fmt.Sprintf("%s/create", app.backendUrl)

		marshal_struct, err := json.Marshal(newLongURL)
		if err != nil {
			app.logger.Errorf("Error marshaling newLongURL: %v", err)
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		client := &http.Client{}
		app.logger.Debugf("Making POST request to: %s with payload: %s", endpointUrl, string(marshal_struct))
		req, err := http.NewRequest("POST", endpointUrl, bytes.NewBuffer(marshal_struct))
		if err != nil {
			app.logger.Println("Error creating POST request:", err.Error())
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			app.logger.Println(err.Error())
			fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			app.logger.Infof("Successfully created short URL for %s", newLongURL.LongUrl)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Propagate backend error back to the client to enable UI alerts
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			app.logger.Errorf("Failed to read backend error response: %v", readErr)
			http.Error(w, "Failed to process backend response", http.StatusInternalServerError)
			return
		}
		// Write same status code and error message
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(res.StatusCode)
		_, _ = w.Write(body)
		return
	}

	// Handle DELETE requests (POST with _method=DELETE)
	if r.Method == http.MethodPost && r.FormValue("_method") == "DELETE" && r.FormValue("delete") == "Delete" {
		shortUrl := r.PostFormValue("ShortUrl")
		app.logger.Debugf("Processing delete request for URL: %s", shortUrl)

		shortCode := shortUrl[len(shortUrl)-7:] // Extract last 7 characters as short code
		endpointUrl := fmt.Sprintf("%s/%s", app.backendUrl, shortCode)
		app.logger.Debugf("Making DELETE request to: %s", endpointUrl)

		client := &http.Client{}
		req, err := http.NewRequest("DELETE", endpointUrl, nil)
		if err != nil {
			app.logger.Errorf("Error creating DELETE request: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			app.logger.Errorf("Error making delete request: %v", err)
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

// health is an HTTP handler for Kubernetes liveness probes.
// A liveness probe checks if the pod is alive and should be restarted if failing.
// This handler only verifies that the frontend service itself is running.
func (app *FrontendApp) health(w http.ResponseWriter, r *http.Request) {
	app.logger.Debug("Liveness probe: frontend is healthy")
	w.WriteHeader(http.StatusOK)
}

// ready is an HTTP handler for Kubernetes readiness probes.
// A readiness probe checks if the pod is ready to accept traffic.
// This handler verifies that the API Gateway backend is accessible and responsive.
func (app *FrontendApp) ready(w http.ResponseWriter, r *http.Request) {
	// Create a client with timeout for readiness check
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Check if backend health endpoint is reachable
	resp, err := client.Get(app.backendUrl + "/health")
	if err != nil {
		app.logger.Warnf("Readiness probe failed: unable to reach backend at %s: %v", app.backendUrl, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Drain and discard the response body to enable connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	// Only consider 2xx status codes as ready
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		app.logger.Debug("Readiness probe: backend is ready")
		w.WriteHeader(http.StatusOK)
	} else {
		app.logger.Warnf("Readiness probe failed: backend returned status %d", resp.StatusCode)
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
