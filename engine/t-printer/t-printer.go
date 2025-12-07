// Package tprinter is for displaying to terminal
package tprinter

type TPrinter interface {
	InitializeVars(vars map[string]any) error
}

type Blocks struct {
	isInline    bool
	keyHandlers map[string]func() error
}

type Page struct {
	blocks []Blocks
}

func New() TPrinter {
	return &tPrinter{}
}

type tPrinter struct {
	vars        map[string]any
	keyHandlers map[string]func() error
	pages       []Page
}

func (tp *tPrinter) InitializeVars(vars map[string]any) error {
	tp.vars = vars

	return nil
}

func (tp *tPrinter) InitializePages(pages []Page) error {
	tp.pages = pages

	return nil
}
