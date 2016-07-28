package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/github"
	keen "github.com/inconshreveable/go-keen"
)

var keenClient = &keen.Client{WriteKey: os.Getenv("KEEN_WRITE_KEY"), ProjectID: os.Getenv("KEEN_PROJECT_ID")}

type event struct {
	Owner       string
	Repo        string
	ReleaseName string
	ReleaseTag  string
	RemoteAddr  string
	URL         string
}

func proxyHandler(client *github.Client, proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			fmt.Fprint(w, "invalid url")
			return
		}

		// part 3 is the architecture and isn't needed
		owner, repo, rest := parts[1], parts[2], parts[4:]
		release, _, err := client.Repositories.GetLatestRelease(owner, repo)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		err = keenClient.AddEvent("usage", &event{
			Owner:       owner,
			Repo:        repo,
			ReleaseName: *release.Name,
			ReleaseTag:  *release.TagName,
			RemoteAddr:  r.RemoteAddr,
			URL:         r.URL.String(),
		})
		if err != nil {
			fmt.Println("keen err:", err)
		}

		restJoin := strings.Join(rest, "/")
		u := fmt.Sprintf("%s/%s/releases/download/%s/%s", owner, repo, *release.TagName, restJoin)

		r.URL.Path = u
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	client := github.NewClient(nil)
	uri, _ := url.Parse("https://github.com")
	proxy := httputil.NewSingleHostReverseProxy(uri)
	http.HandleFunc("/", proxyHandler(client, proxy))

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
