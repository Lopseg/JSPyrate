package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/jquery/esprima"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "js_analyzer",
		Usage: "Analyze JavaScript files to extract URLs, endpoints, and secrets",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input file containing a list of JavaScript files",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "regex",
				Aliases: []string{"r"},
				Usage:   "Optional regex list file for hardcoded secrets",
			},
		},
		Action: func(c *cli.Context) error {
			input := c.String("input")
			regexFile := c.String("regex")

			var regexList []string
			if regexFile != "" {
				regexList = loadRegexList(regexFile)
			}

			jsFiles := loadJSFiles(input)
			var wg sync.WaitGroup
			sem := make(chan bool, 10)

			for _, file := range jsFiles {
				wg.Add(1)
				sem <- true
				go func(filePath string) {
					defer func() {
						<-sem
						wg.Done()
					}()
					analyzeFile(filePath, regexList)
				}(file)
			}
			wg.Wait()
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadJSFiles(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return nil
	}
	defer file.Close()

	var files []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, strings.TrimSpace(scanner.Text()))
	}

	return files
}

func loadRegexList(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filePath, err)
		return nil
	}
	defer file.Close()

	var regexList []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		regexList = append(regexList, strings.TrimSpace(scanner.Text()))
	}

	return regexList
}

func analyzeFile(filePath string, regexList []string) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	content := string(fileContent)

	urls := extractURLs(content)
	endpoints := extractEndpoints(content)

	// Call the parseEndpoints function to extract more complex endpoints
	complexEndpoints := parseEndpoints(content)
	endpoints = append(endpoints, complexEndpoints...)

	secrets := extractSecrets(content, regexList)

	outputFile := fmt.Sprintf("%s_output.txt", strings.TrimSuffix(filePath, ".js"))
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file %s: %v\n", outputFile, err)
		return
	}
	defer output.Close()

	writer := bufio.NewWriter(output)

	writer.WriteString("URLs:\n")
	for _, u := range urls {
		writer.WriteString(u + "\n")
	}

	writer.WriteString("\nEndpoints:\n")
	for _, e := range endpoints {
		writer.WriteString(e + "\n")
	}

	if len(regexList) > 0 {
		writer.WriteString("\nSecrets:\n")
		for _, s := range secrets {
			writer.WriteString(s + "\n")
		}
	}

	writer.Flush()
}

func extractURLs(content string) []string {
	urlRegex := regexp.MustCompile(`https?://[^/\s]+/\S*`)
	return urlRegex.FindAllString(content, -1)
}

func extractEndpoints(content string) []string {
	endpointRegex := regexp.MustCompile(`(["'])\/[\w-]+(?:\/[\w-]+)*\1`)
	matches := endpointRegex.FindAllString(content, -1)

	var endpoints []string
	for _, match := range matches {
		endpoints = append(endpoints, strings.Trim(match, "\"'"))
	}

	return endpoints
}

func parseEndpoints(content string) []string {
	var endpoints []string
	ast, err := esprima.ParseModule(content)
	if err == nil {
		// Implement your logic to parse the AST for complex endpoints
	}

	return endpoints
}

func extractSecrets(content string, regexList []string) []string {
	var secrets []string

	for _, regex := range regexList {
		re := regexp.MustCompile(regex)
		matches := re.FindAllString(content, -1)
		secrets = append(secrets, matches...)
	}

	return secrets
}

