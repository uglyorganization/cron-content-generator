package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	// Read HTML content from a file
	filePath := "file.html"
	htmlContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading HTML file:", err)
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
		postID := div.Attr[0].Val
		author := findAttribute(div, "data-autor")
		// content := extractContent(div)

		fmt.Printf("Post ID: %sAuthor: %s\n", postID, author)
	}
	fmt.Println(len(selectedDivs))
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

func extractContent(n *html.Node) string {
	var content string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			content += n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(n)
	return content
}
