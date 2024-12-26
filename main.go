package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	contentType  = "application/json"
	headerAuth   = "Authorization"
	headerAccept = "Accept"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Website struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type WebsiteStatus struct {
	WebsiteID    int    `json:"websiteId"`
	StatusCode   int    `json:"statusCode"`
	ResponseTime int    `json:"responseTime"`
	PageLoad     int    `json:"pageLoad"`
	PageSize     int64  `json:"pageSize"`
	IsUp         bool   `json:"isUp"`
	CheckedAt    string `json:"checkedAt"`
}

func authenticate(url, email, password string) string {
	loginRequest := LoginRequest{Email: email, Password: password}
	loginRequestBody, _ := json.Marshal(loginRequest)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(loginRequestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(headerAccept, contentType)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("Failed to login, status code: %d, response body: %s", resp.StatusCode, string(body)))
	}

	var loginResponse LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	if err != nil {
		panic(err)
	}

	return loginResponse.Token
}

func fetchWebsites(url, token string) []Website {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set(headerAuth, "Bearer "+token)
	req.Header.Set(headerAccept, contentType)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("Failed to get websites, status code: %d, response body: %s", resp.StatusCode, string(body)))
	}

	var websites []Website
	err = json.NewDecoder(resp.Body).Decode(&websites)
	if err != nil {
		panic(err)
	}

	return websites
}

func checkWebsites(websites []Website) []WebsiteStatus {
	var wg sync.WaitGroup
	statuses := make([]WebsiteStatus, len(websites))
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:      10,
			IdleConnTimeout:   30 * time.Second,
			DisableKeepAlives: false,
		},
	}

	for i, website := range websites {
		wg.Add(1)
		go func(i int, website Website) {
			defer wg.Done()

			start := time.Now()
			resp, err := client.Get(website.URL)
			responseTime := int(time.Since(start).Milliseconds())

			var pageLoad int
			var totalSize int64
			if err == nil {
				pageLoad, totalSize = analyzePage(website.URL)
			}

			status := WebsiteStatus{
				WebsiteID:    website.ID,
				StatusCode:   resp.StatusCode,
				ResponseTime: responseTime,
				PageLoad:     pageLoad,
				PageSize:     totalSize / 1024,
				IsUp:         err == nil && resp.StatusCode == http.StatusOK,
				CheckedAt:    time.Now().Format(time.RFC3339),
			}

			if resp != nil {
				resp.Body.Close()
			}

			statuses[i] = status
		}(i, website)
	}

	wg.Wait()
	return statuses
}

func analyzePage(url string) (int, int64) {
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var html string
	var totalSize int64

	// Starting time for Pageload
	start := time.Now()

	err := chromedp.Run(ctx,
		network.Enable(),
		network.SetBlockedURLS([]string{"*.png", "*.jpg", "*.gif", "*.css", "*.js"}),
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		fmt.Printf("Error analyzing page %s: %v\n", url, err)
		return 0, 0
	}

	pageLoad := int(time.Since(start).Milliseconds())

	totalSize += int64(len(html))
	return pageLoad, totalSize
}

func sendStatuses(url, token string, statuses []WebsiteStatus) {
	statusRequestBody, _ := json.Marshal(statuses)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(statusRequestBody))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(headerAuth, "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		panic(fmt.Sprintf("Failed to send website statuses, status code: %d, body: %s", resp.StatusCode, string(body)))
	}
	fmt.Println("Website statuses sent successfully!")
}

func main() {
	email := os.Getenv("WEBALERT_BACKEND_USER")
	password := os.Getenv("WEBALERT_BACKEND_PASSWORD")
	baseURL := os.Getenv("WEBALERT_BACKEND_LOGIN_URL")
	loginURL := baseURL + "/api/login"

	// Login
	token := authenticate(loginURL, email, password)

	// Get websites
	websitesURL := baseURL + "/api/v1/website/getall"
	websites := fetchWebsites(websitesURL, token)

	// Check websites in parallel
	statuses := checkWebsites(websites)

	// Send statuses
	statusURL := baseURL + "/api/v1/status/setall"
	sendStatuses(statusURL, token, statuses)

	fmt.Println("Website statuses sent successfully!")
}
