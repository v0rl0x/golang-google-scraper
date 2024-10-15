package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const apiKey = "key"
const cx = "key"
const queriesPerMinute = 100
const intervalBetweenQueries = time.Minute / queriesPerMinute

// domains which will be blacklisted in results
var blacklist = []string{
	"github.com",
	"reddit.com",
	"stackexchange.com",
	"stackoverflow.com",
	"quora.com",
	"medium.com",
	"facebook.com",
	"x.com",
	"twitter.com",
	"linkedin.com",
	"pinterest.com",
	"tumblr.com",
	"instagram.com",
	"flickr.com",
	"wikipedia.org",
	"youtube.com",
	"reddit.com",
	"pastebin.com",
	"mozilla.org",
	"duckduckgo.com",
	"sitepoint.com",
	"codecademy.com",
	"bytes.com",
	"programmingforums.org",
	"dev.to",
	"codenewbie.org",
	"slashdot.org",
	"daniweb.com",
	"coderanch.com",
	"gamedev.net",
	"replit.com",
	"community.sap.com",
	"community.spiceworks.com",
	"techguy.org",
	"techsupportforum.com",
	"bleepingcomputer.com/forums",
	"linustechtips.com/main",
	"tomshardware.com/forum",
	"hardforum.com",
	"arstechnica.com/civis",
	"neowin.net/forum",
	"forums.anandtech.com",
  "php.net",
  "microsoft.com",
  "vulnweb.com",
  "intel.com",
}

type SearchResult struct {
    Items []struct {
        Link string `json:"link"`
    } `json:"items"`
    Queries struct {
        NextPage []struct {
            StartIndex int `json:"startIndex"`
        } `json:"nextPage"`
    } `json:"queries"`
}

func isBlacklisted(url string) bool {
	for _, domain := range blacklist {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}

func googleSearch(query, outputFile string) error {
	query = strings.ReplaceAll(query, " ", "+")
	startIndex := 1
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	retryCount := 0
	const maxRetries = 3
	const maxResults = 100
	const queriesPerMinuteLimit = 100
	const waitTime = time.Minute
	queryCount := 0
	fetchedResults := 0
	queryStartTime := time.Now()

	for fetchedResults < maxResults {
		if queryCount >= queriesPerMinuteLimit {
			timeSinceStart := time.Since(queryStartTime)
			if timeSinceStart < time.Minute {
				sleepTime := time.Minute - timeSinceStart
				fmt.Printf("Reached query limit (%d queries/minute). Waiting %s before next query...\n", queriesPerMinuteLimit, sleepTime)
				time.Sleep(sleepTime)
			}
			queryStartTime = time.Now()
			queryCount = 0
		}

		url := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&start=%d&num=10", apiKey, cx, query, startIndex)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch search results: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			retryCount = 0
		} else if resp.StatusCode == http.StatusBadRequest {
			retryCount++
			if retryCount > maxRetries {
				fmt.Printf("Max retries reached for query: %s\n", query)
				break
			}
			fmt.Printf("Retrying query: %s (attempt %d)\n", query, retryCount)
			time.Sleep(2 * time.Second)
			continue
		} else if resp.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Received 429 Too Many Requests, waiting 1 minute...")
			time.Sleep(1 * time.Minute)
			continue
		} else {
			return fmt.Errorf("unexpected response status: %s", resp.Status)
		}

		queryCount++

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		var result SearchResult
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to parse search results: %w", err)
		}

		if len(result.Items) == 0 {
			break
		}

		for _, item := range result.Items {
			if !isBlacklisted(item.Link) {
				if _, err := file.WriteString(item.Link + "\n"); err != nil {
					return fmt.Errorf("failed to write to output file: %w", err)
				}
				fetchedResults++
				if fetchedResults >= maxResults {
					fmt.Printf("Reached the maximum limit of %d results.\n", maxResults)
					return nil
				}
			}
		}

		if len(result.Queries.NextPage) > 0 {
			startIndex = result.Queries.NextPage[0].StartIndex
		} else {
			break
		}

		time.Sleep(intervalBetweenQueries)
	}

	fmt.Printf("Fetched %d results in total.\n", fetchedResults)
	return nil
}

func main() {
	query := flag.String("q", "", "Search query")
	outputFile := flag.String("o", "output.txt", "Output file")
	wordlist := flag.String("w", "", "Wordlist file with multiple search queries")

	flag.Parse()

	if *wordlist != "" {
		file, err := os.Open(*wordlist)
		if err != nil {
			fmt.Printf("Error opening wordlist file: %s\n", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			query := scanner.Text()
			if err := googleSearch(query, *outputFile); err != nil {
				fmt.Printf("Error processing query '%s': %s\n", query, err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading wordlist file: %s\n", err)
			os.Exit(1)
		}
	} else if *query != "" {
		if err := googleSearch(*query, *outputFile); err != nil {
			fmt.Printf("Error processing query '%s': %s\n", *query, err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: go run search.go -q 'search-term' -o output.txt")
		fmt.Println("       go run search.go -w wordlist.txt -o output.txt")
		os.Exit(1)
	}

	fmt.Printf("Search results saved to %s\n", *outputFile)
}
