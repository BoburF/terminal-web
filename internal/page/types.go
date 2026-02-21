package page

import (
	"golang.org/x/net/html"
)

type PageInfo struct {
	Filename    string
	Title       string
	Description string
	Order       int
}

type Page struct {
	Info          PageInfo
	Boxes         []Box
	SectionTitles []string
	Loaded        bool
	Doc           *html.Node
}

type PagePosition struct {
	SectionIndex int
	ScrollOffset int
	LastVisited  int64
}

type Box struct {
	IsPageLink bool
	PageTarget string
	IsNotEmpty bool
	Texts      []string
	Context    []any
}

type PageLink struct {
	Target       string
	SectionTitle string
	Description  string
}
