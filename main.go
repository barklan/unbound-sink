package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const filename = "blacklist.conf"

var urls = []string{
	// StevenBlack blocklist
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	// YouTube ads (kboghdady)
	"https://raw.githubusercontent.com/kboghdady/youTube_ads_4_pi-hole/master/black.list",
	// Ads and tracking extended (lightswitch05)
	"https://raw.githubusercontent.com/lightswitch05/hosts/master/docs/lists/ads-and-tracking-extended.txt",
	// Facebook (lightswitch05)
	"https://raw.githubusercontent.com/lightswitch05/hosts/master/docs/lists/facebook-extended.txt",
	// Tracking aggressive (lightswitch05)
	// "https://raw.githubusercontent.com/lightswitch05/hosts/master/docs/lists/tracking-aggressive-extended.txt",
}

var whiteList = []string{
	"gstaticadssl.l.google.com",
}

func badRegexes() ([]*regexp.Regexp, error) {
	bRgs := []string{
		"^.*#",
		".*localhost.*",
		`\slocal$`,
		`^$`,
	}
	compiled := make([]*regexp.Regexp, len(bRgs))
	for i, rg := range bRgs {
		rgc, err := regexp.Compile(rg)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regexp: %w", err)
		}
		compiled[i] = rgc
	}
	return compiled, nil
}

func parseLine(line string, badRegexes []*regexp.Regexp) (string, bool) {
	for _, rg := range badRegexes {
		if rg.MatchString(line) {
			return "", false
		}
	}
	parsed := strings.Replace(line, "0.0.0.0 ", "", 1)
	if !strings.Contains(parsed, ".") || parsed == "0.0.0.0" {
		return "", false
	}
	fields := strings.Fields(parsed)
	if len(fields) > 1 {
		return "", false
	}
	return parsed, true
}

func unboundLine(domain string) string {
	return `local-zone: "` + domain + "\" always_refuse\n"
}

func parseDoc(input string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var output string
	badRgs, err := badRegexes()
	if err != nil {
		return "", err
	}
	for scanner.Scan() {
		line := scanner.Text()
		parsed, ok := parseLine(line, badRgs)
		if !ok {
			continue
		}
		ready := unboundLine(parsed)
		output += ready
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return output, nil
}

func main() {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	for _, url := range urls {
		r, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			log.Fatalln(r.StatusCode)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalln(err)
		}
		toAppend, err := parseDoc(string(body))
		if err != nil {
			log.Fatalln(err)
		}
		if _, err = f.WriteString(toAppend); err != nil {
			log.Fatalln(err)
		}
	}
}
