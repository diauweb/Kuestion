package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gomarkdown/markdown"

	"github.com/google/go-github/v32/github"
)

type IssueBundle struct {
	Id       int
	Issue    *github.Issue
	Comments []*github.IssueComment
}

func renderText(raw string) interface{} {
	if os.Getenv("BOX_RENDER_MD") == "true" {
		return template.HTML(markdown.ToHTML([]byte(raw), nil, nil))
	}
	return raw
}

var listTmpl = template.Must(
	template.New("").
		Funcs(template.FuncMap{"renderAns": renderText}).
		ParseFiles("tmpl/bakos.gohtml"))
var issues map[int]IssueBundle

func FetchIssues() {
	log.Println("init: get issues")

	rIssues, _, _ := gh.Issues.ListByRepo(ghctx, user, repo,
		&github.IssueListByRepoOptions{
			State:  "all",
			Labels: []string{"publish"},
		})

	issues = make(map[int]IssueBundle)
	for _, v := range rIssues {
		comment, _, _ := gh.Issues.ListComments(ghctx, user, repo, *v.Number, nil)

		issues[*v.Number] = IssueBundle{
			Issue:    v,
			Comments: comment,
		}
	}
}

func listing(w http.ResponseWriter, r *http.Request) {
	err := listTmpl.ExecuteTemplate(w, "bakos.gohtml", issues)
	if err != nil {
		log.Println(err)
	}
}

func trigger(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(os.Getenv("WEBHOOK_SECRET")))
	if err != nil {
		log.Println(err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println(err)
		return
	}

	switch event := event.(type) {
	case *github.IssuesEvent:
		if event.Label == nil || *event.Label.Name != "publish" {
			return
		}

		if *event.Action == "labeled" {
			issue := event.Issue
			comment, _, _ := gh.Issues.ListComments(ghctx, user, repo, *issue.Number, nil)
			issues[*issue.Number] = IssueBundle{
				Issue:    issue,
				Comments: comment,
			}
			log.Println("Issue", *issue.Number, "added")
		} else if *event.Action == "unlabeled" {
			delete(issues, *event.Issue.Number)
			log.Println("Issue", *event.Issue.Number, "removed")
		}
	}
}
