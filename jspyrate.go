package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	urlsFile       string
	secretsFile    string
	outputDir      string
	downloadedPath string
)

func init() {
	flag.StringVar(&urlsFile, "urls", "", "Path to the file containing .js URLs")
	flag.StringVar(&secretsFile, "secrets", "", "Path to the file containing regex patterns for hardcoded secrets (optional)")
	flag.StringVar(&outputDir, "output", "output", "Path to the output directory")
	flag.StringVar(&downloadedPath, "download", "downloaded_js", "Path to the directory where the downloaded .js files will be stored")
}

func main() {
	flag.Parse()

	if urlsFile == "" {
		fmt.Println("Error: missing -urls flag")
		flag.Usage()
		os.Exit(1)
	}

	urls, err := readLines(urlsFile)
	if err != nil {
		fmt.Printf("Error reading URLs file: %v\n", err)
		os.Exit(1)
	}

	secretsRegex := []*regexp.Regexp{}
	if secretsFile != "" {
		secretsPatterns, err := readLines(secretsFile)
		if err != nil {
			fmt.Printf("Error reading secrets file: %v\n", err)
			os.Exit(1)
		}
		for _, pattern := range secretsPatterns {
			secretsRegex = append(secretsRegex, regexp.MustCompile(pattern))
		}
	}

	os.MkdirAll(downloadedPath, os.ModePerm)
	os.MkdirAll(outputDir, os.ModePerm)

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrency to 10 threads

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{} // Acquire a token

		go func(url string) {
			defer func() {
				<-sem // Release the token
				wg.Done()
			}()

			filename := filepath.Base(url)
			downloadJS(url, filepath.Join(downloadedPath, filename))
			endpoints, secrets := analyzeJS(filepath.Join(downloadedPath, filename), secretsRegex)

			writeResults(filepath.Join(outputDir, filename+".txt"), endpoints, secrets)
		}(url)
	}

	wg.Wait()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func downloadJS(url, path string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", path, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error saving file %s: %v\n", path, err)
		return
	}
}

func analyzeJS(path string, secretsRegex []*regexp.Regexp) (endpoints, secrets []string) {
	data,	err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
		return
	}

	content := string(data)

	endpointRegex := regexp.MustCompile(`\b(?:https?://|/)[^"'\\]*(?:\.json|\.php|\.aspx?|\.jsp|\.do|\.cgi|\.cfm|\.action)[^"'\\]*\b`)
	urlRegex := regexp.MustCompile(`\bhttps?://[^\s"'\\]+(?:\.html|\.htm|\.js|\.css|\.xml|\.rss|\.atom|\.txt)[^"'\\]*\b`)

	endpointMatches := endpointRegex.FindAllString(content, -1)
	urlMatches := urlRegex.FindAllString(content, -1)

	endpoints = append(endpoints, endpointMatches...)
	endpoints = append(endpoints, urlMatches...)

	for _, re := range secretsRegex {
		secretsMatches := re.FindAllString(content, -1)
		secrets = append(secrets, secretsMatches...)
	}

	return
}

func writeResults(path string, endpoints, secrets []string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating output file %s: %v\n", path, err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, endpoint := range endpoints {
		writer.WriteString("Endpoint: " + endpoint + "\n")
	}

	if len(secrets) > 0 {
		writer.WriteString("\nHardcoded secrets:\n")
		for _, secret := range secrets {
			writer.WriteString(secret + "\n")
		}
	}
}


