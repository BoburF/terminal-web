package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Box struct {
	isNotEmplty bool
	texts       []string
	inputs      []textinput.Model
	context     []any
}

type State struct {
	Width  int
	Height int
	boxes  []Box
}

func (s State) Init() tea.Cmd {
	return nil
}

func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		}
	}

	return s, cmd
}

func (s State) View() string {
	var outputStr strings.Builder

	for _, box := range s.boxes {
		if !box.isNotEmplty {
			continue
		}

		boxStyle := lipgloss.NewStyle().Border(lipgloss.BlockBorder()).Padding(1).Width(s.Width).Align(lipgloss.Center)
		var boxStr strings.Builder
		for _, context := range box.context {
			switch ctx := context.(type) {
			case string:
				boxStr.WriteString(lipgloss.NewStyle().Bold(true).Render(ctx) + "\n")
			case textinput.Model:
				boxStr.WriteString(ctx.View() + "\n")
			}
		}

		outputStr.WriteString(boxStyle.Render(boxStr.String()) + "\n")
	}

	return outputStr.String()
}
