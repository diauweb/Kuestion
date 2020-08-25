package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/joho/godotenv"
	"github.com/kataras/hcaptcha"
	"golang.org/x/oauth2"
)

var client *hcaptcha.Client
var gh *github.Client
var ghctx = context.Background()

var user, repo string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client = hcaptcha.New(os.Getenv("HCAPTCHA_SECRET_KEY"))

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_PAT")},
	)
	gh = github.NewClient(oauth2.NewClient(ghctx, ts))
	rn := strings.Split(os.Getenv("GH_REPO"), "/")
	user, repo = rn[0], rn[1]

	FetchIssues()

	http.HandleFunc("/sbmt", submit)
	http.HandleFunc("/box", listing)
	http.HandleFunc("/trigger", trigger)

	if os.Getenv("STANDALONE") == "true" {
		log.Println("Standalone mode")
		http.Handle("/", http.FileServer(http.Dir("./static")))
	}

	log.Printf("Listening on: localhost:%s\n", os.Getenv("PORT"))

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

func postIssue(body string) error {

	msg := fmt.Sprintf("Bako sent at %s", time.Now().Format("2006-01-02 15:04:05 +0800"))
	_, _, err := gh.Issues.Create(ghctx, user, repo,
		&github.IssueRequest{
			Title: &msg,
			Body:  &body,
		})

	return err
}
