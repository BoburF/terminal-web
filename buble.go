package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Box struct {
	texts  []string
	inputs []textinput.Model
}

type State struct {
	boxes    []Box
	inputs   []textinput.Model
	buttons  []string
	handlers map[string]func() int
	focused  int
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
		handler := s.handlers[msg.String()]
		if handler != nil {
			handler()
		}

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
		for _, text := range box.texts {
			b.WriteString(text)
			b.WriteString("\n")
		}

		for i, input := range box.inputs {
			if i == s.focused {
				input.Focus()
			} else {
				input.Blur()
			}
			b.WriteString(input.View())
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(s.buttons) > 0 {
		b.WriteString("\n")
		for _, button := range s.buttons {
			b.WriteString(button)
			b.WriteString(" ")
		}
		b.WriteString("\n")
	}

	return b.String()
}
