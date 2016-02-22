package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

func proxyHandler(client *github.Client, proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			fmt.Fprint(w, "invalid url")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// part 3 is the architecture and isn't needed
		owner, repo, rest := parts[1], parts[2], parts[4:]
		release, _, err := client.Repositories.GetLatestRelease(owner, repo)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		restJoin := strings.Join(rest, "/")
		u := fmt.Sprintf("%s/%s/releases/download/%s/%s", owner, repo, *release.TagName, restJoin)
		log.Printf("%s %s %s --> %s", r.RemoteAddr, r.Method, r.URL, u)

		r.URL.Path = u
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	log.SetPrefix("apk2gh | ")

	client := github.NewClient(nil)
	uri, _ := url.Parse("https://github.com")
	proxy := httputil.NewSingleHostReverseProxy(uri)

	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+port, proxyHandler(client, proxy)))
}
