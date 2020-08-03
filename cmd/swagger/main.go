package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
)

type GitHubReleaseAsset struct {
	Url                string `json:"url"`
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type GithubReleaseResponse struct {
	Assets []GitHubReleaseAsset `json:"assets"`
}

func main() {
	url := "https://api.github.com/repos/go-swagger/go-swagger/releases/latest"

	client := &http.Client{}

	jsonRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if os.Getenv("GITHUB_TOKEN") != "" {
		jsonRequest.Header.Add("authorization", "bearer "+os.Getenv("GITHUB_TOKEN"))
	}

	jsonResponse, err := client.Do(jsonRequest)
	if err != nil {
		log.Fatalf("failed to download information about the latest go-swagger release (%v)", err)
	}
	defer jsonResponse.Body.Close()

	if jsonResponse.StatusCode != 200 {
		log.Fatalf("invalid HTTP response code for release query (%s)", jsonResponse.Status)
	}

	releaseResponse := &GithubReleaseResponse{}
	err = json.NewDecoder(jsonResponse.Body).Decode(releaseResponse)
	if err != nil {
		log.Fatalf("failed to decode github release data (%v)", err)
	}

	var file string
	if runtime.GOOS == "windows" {
		file = "swagger_" + runtime.GOOS + "_" + runtime.GOARCH + ".exe"
	} else {
		file = "swagger_" + runtime.GOOS + "_" + runtime.GOARCH
	}
	url = ""
	for _, asset := range releaseResponse.Assets {
		log.Printf("Asset: %s | %s", asset.Name, asset.BrowserDownloadUrl)
		if asset.Name == file {
			url = asset.BrowserDownloadUrl
		}
	}
	if url == "" {
		log.Fatalf("failed to find URL for go-swagger executable")
	}

	tempDir := os.TempDir()
	fullPath := path.Join(tempDir, file)

	binaryRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if os.Getenv("GITHUB_TOKEN") != "" {
		jsonRequest.Header.Add("authorization", "bearer "+os.Getenv("GITHUB_TOKEN"))
	}
	binaryResponse, err := client.Do(binaryRequest)
	if err != nil {
		log.Fatalf("failed to download information about the latest go-swagger release (%v)", err)
	}
	defer binaryResponse.Body.Close()

	binaryData, err := ioutil.ReadAll(binaryResponse.Body)
	if err != nil {
		log.Fatalf("failed to read downloaded file (%v)", err)
	}

	err = ioutil.WriteFile(fullPath, binaryData, 0755)
	if err != nil {
		log.Fatalf("failed to write executable binary (%v)", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed toget cwd (%v)", err)
	}

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		cmd := exec.Command(fullPath, "generate", "spec", "-mo", arg)
		cmd.Dir = cwd
		err = cmd.Run()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}
