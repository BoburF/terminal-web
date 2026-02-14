package main

import (
	"fmt"
	"strconv"
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
	Width               int
	Height              int
	boxes               []Box
	interactivity       []Controller
	quitting            bool
	session             any
	currentSection      int
	sectionScrollOffset int
	pendingSectionNum   string
}

func (s State) Init() tea.Cmd {
	return nil
}

func (s State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		s.Height = msg.Height
		return s, nil

	case tea.KeyMsg:
		keyStr := msg.String()

		if len(keyStr) == 1 && keyStr >= "0" && keyStr <= "9" {
			s.pendingSectionNum += keyStr
			return s, nil
		}

		if keyStr == "esc" {
			s.pendingSectionNum = ""
			return s, nil
		}

		switch keyStr {
		case "tab":
			if s.pendingSectionNum != "" {
				sectionNum, _ := strconv.Atoi(s.pendingSectionNum)
				s.pendingSectionNum = ""
				if sectionNum > 0 {
					s.currentSection = sectionNum - 1
					if s.currentSection >= len(s.boxes) {
						s.currentSection = len(s.boxes) - 1
					}
					s.sectionScrollOffset = 0
				}
			} else {
				s.currentSection++
				if s.currentSection >= len(s.boxes) {
					s.currentSection = 0
				}
				s.sectionScrollOffset = 0
			}
			return s, nil
		case "shift+tab":
			s.currentSection--
			if s.currentSection < 0 {
				s.currentSection = len(s.boxes) - 1
			}
			s.sectionScrollOffset = 0
			s.pendingSectionNum = ""
			return s, nil
		case "j", "down":
			if s.shouldScrollInternally() {
				s.sectionScrollOffset++
				maxOffset := s.getMaxScrollOffset()
				if s.sectionScrollOffset > maxOffset {
					s.sectionScrollOffset = maxOffset
				}
			}
			s.pendingSectionNum = ""
			return s, nil
		case "k", "up":
			if s.shouldScrollInternally() {
				s.sectionScrollOffset--
				if s.sectionScrollOffset < 0 {
					s.sectionScrollOffset = 0
				}
			}
			s.pendingSectionNum = ""
			return s, nil
		case "ctrl+c", "q":
			s.quitting = true
			return s, tea.Quit
		}

		for _, ctrl := range s.interactivity {
			if ctrl.combination == keyStr {
				return s.handleController(ctrl)
			}
		}
	}

	return s, nil
}

func (s State) handleController(ctrl Controller) (tea.Model, tea.Cmd) {
	switch ctrl.event {
	case "exit":
		s.quitting = true
		return s, tea.Quit
	case "switch-sections":
		s.currentSection++
		if s.currentSection >= len(s.boxes) {
			s.currentSection = 0
		}
		s.sectionScrollOffset = 0
		return s, nil
	default:
		return s, nil
	}
}

func (s State) getContentHeight() int {
	return s.Height - 6
}

func (s State) getSectionHeight(sectionIdx int) int {
	if sectionIdx < 0 || sectionIdx >= len(s.boxes) {
		return 0
	}
	box := s.boxes[sectionIdx]
	lineCount := 0
	for _, ctx := range box.context {
		switch content := ctx.(type) {
		case string:
			lineCount += strings.Count(content, "\n") + 1
		case *textinput.Model:
			lineCount += 1
		}
	}
	return lineCount
}

func (s State) shouldScrollInternally() bool {
	return s.getSectionHeight(s.currentSection) > s.getContentHeight()
}

func (s State) getMaxScrollOffset() int {
	return s.getSectionHeight(s.currentSection) - s.getContentHeight()
}

func (s State) View() string {
	if s.quitting {
		return ""
	}

	if s.currentSection >= len(s.boxes) {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(s.Width - 2).
		Height(s.getContentHeight() + 2).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	indicatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Bold(true).
		Width(s.Width - 2).
		AlignHorizontal(lipgloss.Center)

	controllerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true)

	bindingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)

	pendingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	box := s.boxes[s.currentSection]

	var contentLines []string
	for _, ctx := range box.context {
		switch content := ctx.(type) {
		case string:
			contentLines = append(contentLines, strings.Split(content, "\n")...)
		case *textinput.Model:
			contentLines = append(contentLines, content.View())
		}
	}

	visibleHeight := s.getContentHeight()
	if len(contentLines) > visibleHeight {
		endIdx := s.sectionScrollOffset + visibleHeight
		if endIdx > len(contentLines) {
			endIdx = len(contentLines)
		}
		contentLines = contentLines[s.sectionScrollOffset:endIdx]
	}

	sectionNum := s.currentSection + 1
	totalSections := len(s.boxes)

	var indicatorText string
	if s.pendingSectionNum != "" {
		indicatorText = pendingStyle.Render(fmt.Sprintf("%d/%d: Jumping to section %s...", sectionNum, totalSections, s.pendingSectionNum))
	} else {
		indicatorText = indicatorStyle.Render(fmt.Sprintf("%d/%d: %s - Press Tab to switch", sectionNum, totalSections, getSectionTitle(box)))
	}

	var controllerParts []string
	for _, controller := range s.interactivity {
		name := controllerStyle.Render(controller.name)
		binding := bindingStyle.Render(fmt.Sprintf("[%s]", controller.combination))
		controllerParts = append(controllerParts, fmt.Sprintf("%s %s", name, binding))
	}

	controllerLine := strings.Join(controllerParts, "  ")

	result := indicatorText + "\n" + boxStyle.Render(strings.Join(contentLines, "\n")) + "\n" + controllerLine

	return result
}

func getSectionTitle(box Box) string {
	for _, ctx := range box.context {
		if str, ok := ctx.(string); ok {
			return str
		}
	}
	return "Section"
}
