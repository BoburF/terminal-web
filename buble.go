package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type State struct {
	boxes   []string
	inputs  []textinput.Model
	buttons []string
	focused int
}

func (s State) Init() tea.Cmd {
	if len(s.inputs) > 0 {
		return s.inputs[0].Focus()
	}
	return nil
}

func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		case "tab":
			s.focused = (s.focused + 1) % len(s.inputs)
			for i := range s.inputs {
				if i == s.focused {
					cmd = s.inputs[i].Focus()
				} else {
					s.inputs[i].Blur()
				}
			}
		}
	}

	for i := range s.inputs {
		s.inputs[i], cmd = s.inputs[i].Update(msg)
	}

	return s, cmd
}

func (s State) View() string {
	var b strings.Builder

	for _, box := range s.boxes {
		b.WriteString(box)
		b.WriteString("\n")
	}

	for i, input := range s.inputs {
		if i == s.focused {
			input.Focus()
		} else {
			input.Blur()
		}
		b.WriteString(input.View())
		b.WriteString("\n")
	}

	if len(s.buttons) > 0 {
		b.WriteString("\n")
		for _, button := range s.buttons {
			b.WriteString(button)
			b.WriteString(" ")
		}
	}

	return b.String()
}
