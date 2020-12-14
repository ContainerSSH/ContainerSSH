package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

func main() {
	cmd := exec.Command("golicense", os.Args[1], os.Args[2])
	var stdOutBuf bytes.Buffer
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = os.Stderr
	cmd.Env = nil
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to run golicense (%v)", err)
	}
	data := stdOutBuf.String()
	lines := strings.Split(data, "\n")
	noticeFiles := map[string]string{}
	licenses := map[string]string{}
	re := regexp.MustCompile(`\s+`)

	cmd = exec.Command("go", "mod", "vendor")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("%v", err)
	}

	extractLicenses(lines, re, licenses, noticeFiles)

	noticeContents := renderNotice(licenses, noticeFiles)

	err = ioutil.WriteFile(os.Args[3], noticeContents, 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = os.RemoveAll("vendor")
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func extractLicenses(lines []string, re *regexp.Regexp, licenses map[string]string, noticeFiles map[string]string) {
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := re.Split(line, -1)
		if len(parts) < 2 {
			continue
		}
		repo := strings.TrimSpace(parts[0])
		licenses[repo] = strings.TrimSpace(parts[1])
		noticeFile := path.Join("vendor", repo, "NOTICE")
		noticeFileMarkdown := path.Join("vendor", repo, "NOTICE.md")
		if _, err := os.Stat(noticeFile); err == nil {
			noticeFiles[repo] = noticeFile
		} else if _, err := os.Stat(noticeFileMarkdown); err == nil {
			noticeFiles[repo] = noticeFileMarkdown
		}
	}
}

func renderNotice(licenses map[string]string, noticeFiles map[string]string) []byte {
	var finalNotice bytes.Buffer
	finalNotice.Write([]byte("# Third party licenses\n\n"))
	finalNotice.Write([]byte("This project contains third party packages under the following licenses:\n\n"))
	for packageName, license := range licenses {
		finalNotice.Write([]byte(fmt.Sprintf("## [%s](https://%s)\n\n", packageName, packageName)))
		finalNotice.Write([]byte(fmt.Sprintf("**License:** %s\n\n", license)))

		if noticeFile, ok := noticeFiles[packageName]; ok {
			noticeData, err := ioutil.ReadFile(noticeFile)
			if err != nil {
				log.Fatalf("failed to read notice file %s (%v)", noticeFile, err)
			}
			trimmedNotice := strings.TrimSpace(string(noticeData))
			if trimmedNotice != "" {
				finalNotice.Write([]byte(fmt.Sprintf("%s\n\n", "> "+strings.ReplaceAll(trimmedNotice, "\n", "\n> "))))
			}
		}
	}
	noticeContents := finalNotice.Bytes()
	return noticeContents
}
