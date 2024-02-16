package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v57/github"
)

func TestFollowRedirects(t *testing.T) {
	ctx := context.Background()

	httpClient := &http.Client{
		Transport: withLogging(t, http.DefaultTransport),
	}

	gh := github.NewClient(httpClient)

	const (
		owner = "mterwill"
		repo  = "go-github-issue-demo" // renamed to go-github-issue-demo-1
	)

	releases, _, err := gh.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		t.Fatalf("listing releases: %s", err)
	}
	var assetID int64
	for _, release := range releases {
		for _, asset := range release.Assets {
			if asset.GetName() == "foo.txt" {
				assetID = asset.GetID()
				break
			}
		}
	}
	if assetID == 0 {
		t.Fatalf("no assets found")
	}

	reader, _, err := gh.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID, httpClient)
	if err != nil {
		t.Fatalf("downloading asset: %s", err)
	}
	defer reader.Close()

	gotData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading: %s", err)
	}
	wantData := "Hello, world!\n"
	if diff := cmp.Diff(string(gotData), wantData); diff != "" {
		t.Fatalf("data differs (-got +want):\n%s", diff)
	}
}

func withLogging(t *testing.T, base http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (res *http.Response, _ error) {
		raw, _ := httputil.DumpRequest(req, false)
		t.Logf("Request: %s", raw)
		defer func() {
			raw, _ := httputil.DumpResponse(res, false)
			t.Logf("Response: %s", raw)
		}()
		return base.RoundTrip(req)
	})
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
