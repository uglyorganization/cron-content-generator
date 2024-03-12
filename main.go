package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type post struct {
	id    string
	votes int
}

type authorInfo struct {
	totalVotes int
	posts      []post
}

type authorPair struct {
	User        string `json:"user"`
	TotalVotes  int    `json:"totalVotes"`
	NumberPosts int    `json:"numberPosts"`
	MaxVotes    int    `json:"maxVotes"`
}

func main() {
	authors := make(map[string]*authorInfo)
	pages := 1798

	for i := 0; i < pages+1; i++ {
		log.Printf("iteration: %d\n", i)
		time.Sleep(220 * time.Millisecond)
		baseURL := "https://www.mediavida.com/foro/dev/feda-dev-no-javascript-allowed-643822"
		url := fmt.Sprintf("%s/%d", baseURL, i)
		htmlContent, err := getHTMLContent(url)
		if err != nil {
			log.Fatal("Error getting HTML:", err)
			return
		}

		doc, err := html.Parse(strings.NewReader(string(htmlContent)))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}

		// Select all divs with the specified structure
		selectedDivs := findDivs(doc, "post")
		for _, div := range selectedDivs {
			postID := findNumericPostID(div.Attr)
			author := findAttribute(div, "data-autor")
			votes := findVotes(div)
			if votes == 0 {
				continue
			}

			post := post{id: postID, votes: votes}
			log.Println(postID)

			info, exists := authors[author]
			if !exists {
				info = &authorInfo{}
				authors[author] = info
			}

			info.posts = append(info.posts, post)
			info.totalVotes += post.votes
		}

	}

	var keyValueList []authorPair
	for author, info := range authors {
		maxVotes := 0
		for _, post := range info.posts {
			if post.votes > maxVotes {
				maxVotes = post.votes
			}
		}
		authorPair := authorPair{User: author, TotalVotes: info.totalVotes, NumberPosts: len(info.posts), MaxVotes: maxVotes}
		keyValueList = append(keyValueList, authorPair)
	}

	// Step 1: Open a file for writing
	file, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Step 2: Marshal the list of structs to JSON
	jsonData, err := json.Marshal(keyValueList)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Step 3: Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

}

func getHTMLContent(url string) (string, error) {
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

func findVotes(n *html.Node) int {
	var votes int

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "i" && hasClass(n, "fa-thumbs-up") {
			for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
				if sibling.Type == html.ElementNode && sibling.Data == "span" {
					votesStr := strings.TrimSpace(sibling.FirstChild.Data)
					parsedVotes, err := strconv.Atoi(votesStr)
					if err == nil {
						votes = parsedVotes
					}
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return votes
}

func findDivs(n *html.Node, className string) map[string]*html.Node {
	divs := make(map[string]*html.Node)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, className) {
			postID := findNumericPostID(n.Attr)
			if postID != "" {
				divs[postID] = n
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return divs
}

func findNumericPostID(attrs []html.Attribute) string {
	for _, attr := range attrs {
		if attr.Key == "id" && strings.HasPrefix(attr.Val, "post-") {
			return strings.TrimPrefix(attr.Val, "post-")
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			return true
		}
	}

	return false
}

func findAttribute(n *html.Node, attributeName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attributeName {
			return attr.Val
		}
	}

	return ""
}
