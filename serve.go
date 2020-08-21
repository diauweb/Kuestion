package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kataras/hcaptcha"
)

var client *hcaptcha.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client = hcaptcha.New(os.Getenv("HCAPTCHA_SECRET_KEY"))

	http.HandleFunc("/sbmt", submit)
	// http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Printf("Listening on: localhost:%s\n", os.Getenv("PORT"))

	_ = http.ListenAndServe("localhost:"+os.Getenv("PORT"), nil)
}

func submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Println(err)
	}

	verify := client.SiteVerify(r)
	if !verify.Success {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Error while validating your request. %#+v", verify)
		return
	}

	text := r.Form.Get("text")

	if text == "" {
		fmt.Fprintf(w, "The text field is empty in your request.")
		w.WriteHeader(400)
		return
	}

	err = postIssue(text)

	if err != nil {
		fmt.Fprintf(w, "Internal Server Error.")
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Location", os.Getenv("SUCCESS_CALLBACK")+"?ok=1")
	w.WriteHeader(303)
}

const issueApi = "https://api.github.com/repos/%s/issues"

func postIssue(body string) error {
	payload, err := json.Marshal(map[string]string{
		"title": fmt.Sprintf("Bako sent at %s", time.Now()),
		"body": body,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(issueApi, os.Getenv("GH_REPO")),
		bytes.NewBuffer(payload),
	)

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+os.Getenv("GH_PAT"))
	req.Header.Set("Content-Type", "application/json")

	_, err = http.DefaultClient.Do(req)

	return err
}
