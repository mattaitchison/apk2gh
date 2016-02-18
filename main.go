package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/mattaitchison/apkghproxy/Godeps/_workspace/src/github.com/google/go-github/github"
)

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
