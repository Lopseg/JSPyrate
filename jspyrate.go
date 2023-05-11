package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func extractEndpoints(data string, regexes []string) []string {
	var endpoints []string

	for _, regexStr := range regexes {
		re := regexp.MustCompile(regexStr)
		matches := re.FindAllString(data, -1)
		endpoints = append(endpoints, matches...)
	}

	return endpoints
}

func loadRegexes(wordlistPath string) ([]string, error) {
	data, err := ioutil.ReadFile(wordlistPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	return lines, nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func processURL(url string, outputPath string, regexes []string, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := downloadFile(url)
	if err != nil {
		fmt.Printf("Error downloading file: %s\n", url)
		return
	}

	content := data
	endpoints := extractEndpoints(content, regexes)

	if len(endpoints) > 0 {
		output := strings.Join(endpoints, "\n")

		filename := filepath.Base(url) + ".txt"
		outFile := filepath.Join(outputPath, filename)
		err := ioutil.WriteFile(outFile, []byte(output), 0644)
		if err != nil {
			fmt.Printf("Error writing to file: %s\n", outFile)
		}
	}
}

func main() {
	var wordlistPath string
	var inputURL string
	var outputPath string
	var threads int

	flag.StringVar(&wordlistPath, "wordlist", "", "Path to the wordlist containing regexes for hardcoded secrets")
	flag.StringVar(&inputURL, "url", "", "A single URL of a JavaScript file")
	flag.StringVar(&outputPath, "output", "", "Path to the directory to save the output files")
	flag.IntVar(&threads, "threads", 10, "Number of concurrent threads")

	flag.Parse()

	if outputPath == "" {
		fmt.Println("The output directory must be specified")
		os.Exit(1)
	}

	var regexes []string
	var err error

	if wordlistPath != "" {
		regexes, err = loadRegexes(wordlistPath)
		if err != nil {
			fmt.Printf("Error reading wordlist: %s\n", wordlistPath)
			os.Exit(1)
		}
	}

	var wg sync.WaitGroup
	sem := make(chan bool, threads)

	if inputURL != "" {
		wg.Add(1)
		sem <- true
		go func(url string) {
			processURL(url, outputPath, regexes, &wg)
			<-sem
		}(inputURL)
	}

	wg.Wait()
}
