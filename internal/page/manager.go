package page

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var RootPath = "./resume/"

func DiscoverPages(rootPath string) ([]PageInfo, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	var pages []PageInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		info, err := extractPageInfo(filepath.Join(rootPath, entry.Name()))
		if err != nil {
			continue
		}

		info.Filename = entry.Name()
		pages = append(pages, info)
	}

	sort.Slice(pages, func(i, j int) bool {
		if pages[i].Order != 0 || pages[j].Order != 0 {
			return pages[i].Order < pages[j].Order
		}
		return pages[i].Filename < pages[j].Filename
	})

	return pages, nil
}

func extractPageInfo(filePath string) (PageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return PageInfo{}, err
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return PageInfo{}, err
	}

	info := PageInfo{}

	var findBody func(*html.Node)
	findBody = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			for _, attr := range node.Attr {
				switch attr.Key {
				case "page-title":
					info.Title = attr.Val
				case "page-desc", "page-description":
					info.Description = attr.Val
				case "page-order":
					info.Order, _ = strconv.Atoi(attr.Val)
				}
			}
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findBody(child)
		}
	}

	findBody(doc)

	if info.Title == "" {
		info.Title = strings.TrimSuffix(filepath.Base(filePath), ".html")
		info.Title = strings.ReplaceAll(info.Title, "-", " ")
		info.Title = strings.Title(info.Title)
	}

	return info, nil
}

func LoadPage(filename string) (*html.Node, error) {
	filePath := filepath.Join(RootPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return nil, err
	}

	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			body = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findBody(child)
		}
	}

	findBody(doc)

	return body, nil
}

func GetPageTitle(doc *html.Node) string {
	title := ""
	var findTitle func(*html.Node)
	findTitle = func(node *html.Node) {
		for _, attr := range node.Attr {
			if attr.Key == "page-title" {
				title = attr.Val
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findTitle(child)
		}
	}
	findTitle(doc)
	return title
}
