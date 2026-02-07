package main

import (
	"fmt"
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

type Controller struct {
	event       string
	combination string
	name        string
}

type State struct {
	Width         int
	Height        int
	boxes         []Box
	interactivity []Controller
	quitting      bool
}

func (s State) Init() tea.Cmd {
	return nil
}

func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			s.quitting = true
			return s, tea.Quit
		}
	}

	return s, nil
}

func (s State) View() string {
	if s.quitting {
		return ""
	}

	var allContent []string
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(s.Width - 2).
		Height(s.Height - 2).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	controllerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true)

	bindingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)

	for _, box := range s.boxes {
		if !box.isNotEmplty {
			continue
		}

		boxStr := make([]string, 0)
		for _, context := range box.context {
			switch ctx := context.(type) {
			case string:
				boxStr = append(boxStr, ctx)
			case *textinput.Model:
				boxStr = append(boxStr, ctx.View())
			}
		}

		allContent = append(allContent, strings.Join(boxStr, "\n"))
	}

	var controllerParts []string
	for _, controller := range s.interactivity {
		name := controllerStyle.Render(controller.name)
		binding := bindingStyle.Render(fmt.Sprintf("[%s]", controller.combination))
		controllerParts = append(controllerParts, fmt.Sprintf("%s %s", name, binding))
	}

	availableHeight := s.Height - 6

	contentHeight := 0
	for _, content := range allContent {
		contentHeight += strings.Count(content, "\n") + 1
	}

	spacing := availableHeight - contentHeight
	if spacing > 0 && len(controllerParts) > 0 {
		for range spacing {
			allContent = append(allContent, "")
		}
	}

	if len(controllerParts) > 0 {
		allContent = append(allContent, strings.Join(controllerParts, "  "))
	}

	return boxStyle.Render(strings.Join(allContent, "\n"))
}
