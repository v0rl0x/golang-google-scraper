# Google Custom Search Scraper

Simple scraper used to search google for terms and output the URLs. Useful for SQLI searching and CVE dorks.

## Prerequisites

- Go 1.16 or later
- Google Custom Search API Key
- Google Custom Search Engine ID (cx)

### Options

- `-q`: A single search query without the use of a file.
- `-o`: Output file where results will be saved.
- `-w`: Wordlist file containing multiple search queries. One query per line such as "inurl: hello"

### How to use

go run search.go -q "search-term" -o output.txt

go run search.go -w file.txt -o output.txt
