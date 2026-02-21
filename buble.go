package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BoburF/terminal-web.git/internal/page"
)

type Box struct {
	isNotEmplty bool
	texts       []string
	inputs      []textinput.Model
	context     []any
	IsPageLink  bool
	PageTarget  string
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
	sectionTitles       []string
	pages               []page.PageInfo
	currentPageIdx      int
	pageHistory         []int
	showPagePrompt      bool
	pendingPageIdx      int
	promptMessage       string
	notSwitchedMsg      string
	notSwitchedTimer    int
	pageLinks           []page.PageLink
	lastTabPressed      bool
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
			if s.showPagePrompt {
				s.showPagePrompt = false
				s.notSwitchedMsg = "Not switched"
				s.notSwitchedTimer = 3
				s.currentSection++
				if s.currentSection >= len(s.boxes) {
					s.currentSection = 0
				}
				s.sectionScrollOffset = 0
				return s, nil
			}
			return s, nil
		}

		if s.showPagePrompt {
			switch keyStr {
			case "enter":
				s.lastTabPressed = false
				return s.confirmPageSwitch()
			case "y":
				s.lastTabPressed = false
				return s.confirmPageSwitch()
			case "n":
				s.showPagePrompt = false
				s.notSwitchedMsg = "Not switched"
				s.notSwitchedTimer = 3
				s.currentSection++
				if s.currentSection >= len(s.boxes) {
					s.currentSection = 0
				}
				s.sectionScrollOffset = 0
				s.lastTabPressed = false
				return s, nil
			}
			return s, nil
		}

		if keyStr == "enter" && s.lastTabPressed {
			s.lastTabPressed = false
			if s.currentSection >= 0 && s.currentSection < len(s.pageLinks) {
				link := s.pageLinks[s.currentSection]
				if link.Target != "" {
					pageIdx := s.findPageIndex(link.Target)
					if pageIdx >= 0 {
						s.pendingPageIdx = pageIdx
						s.promptMessage = link.SectionTitle
						return s.confirmPageSwitch()
					}
				}
			}
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
			s.lastTabPressed = true
			return s, nil
		case "shift+tab":
			s.currentSection--
			if s.currentSection < 0 {
				s.currentSection = len(s.boxes) - 1
			}
			s.sectionScrollOffset = 0
			s.pendingSectionNum = ""
			s.lastTabPressed = false
			return s, nil
		case "j", "down":
			s.lastTabPressed = false
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
			s.lastTabPressed = false
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
	case "back":
		return s.navigateBack()
	default:
		return s, nil
	}
}

func (s State) confirmPageSwitch() (State, tea.Cmd) {
	if s.pendingPageIdx >= 0 && s.pendingPageIdx < len(s.pages) {
		pageInfo := s.pages[s.pendingPageIdx]

		body, err := page.LoadPage(pageInfo.Filename)
		if err != nil {
			s.showPagePrompt = false
			s.notSwitchedMsg = "Error loading page"
			s.notSwitchedTimer = 3
			return s, nil
		}

		boxes, sectionTitles, pageLinks := parseMain(body)
		controllers := parseControlls(body)

		s.pageHistory = append(s.pageHistory, s.currentPageIdx)
		s.currentPageIdx = s.pendingPageIdx
		s.boxes = boxes
		s.sectionTitles = sectionTitles
		s.pageLinks = pageLinks
		s.interactivity = controllers
		s.currentSection = 0
		s.sectionScrollOffset = 0
		s.showPagePrompt = false
		s.pendingPageIdx = 0
		s.promptMessage = ""
	}
	return s, nil
}

func (s State) navigateBack() (State, tea.Cmd) {
	if len(s.pageHistory) > 0 {
		lastIdx := len(s.pageHistory) - 1
		prevPageIdx := s.pageHistory[lastIdx]
		s.pageHistory = s.pageHistory[:lastIdx]

		if prevPageIdx >= 0 && prevPageIdx < len(s.pages) {
			pageInfo := s.pages[prevPageIdx]
			body, err := page.LoadPage(pageInfo.Filename)
			if err == nil {
				boxes, sectionTitles, pageLinks := parseMain(body)
				controllers := parseControlls(body)

				s.boxes = boxes
				s.sectionTitles = sectionTitles
				s.pageLinks = pageLinks
				s.interactivity = controllers
			}
		}

		s.currentPageIdx = prevPageIdx
		s.currentSection = 0
		s.sectionScrollOffset = 0
	}
	return s, nil
}

func (s State) checkPageLinkSection() {
	if len(s.pageLinks) == 0 {
		return
	}
	if s.currentSection >= 0 && s.currentSection < len(s.pageLinks) {
		link := s.pageLinks[s.currentSection]
		if link.Target != "" {
			s.showPagePrompt = true
			s.pendingPageIdx = s.findPageIndex(link.Target)
			s.promptMessage = link.SectionTitle
		}
	}
}

func (s State) findPageIndex(filename string) int {
	for i, p := range s.pages {
		if p.Filename == filename {
			return i
		}
	}
	return -1
}

func (s State) renderPagePrompt() string {
	promptWidth := 50
	promptHeight := 8

	x := (s.Width - promptWidth) / 2
	y := (s.Height - promptHeight) / 2

	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}

	promptStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(promptWidth).
		Height(promptHeight).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Bold(true).
		Width(promptWidth - 4).
		AlignHorizontal(lipgloss.Center)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Width(promptWidth - 4).
		AlignHorizontal(lipgloss.Center)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true)

	skipKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)

	pageTitle := s.promptMessage
	if s.pendingPageIdx >= 0 && s.pendingPageIdx < len(s.pages) {
		pageTitle = s.pages[s.pendingPageIdx].Title
	}

	promptContent := titleStyle.Render("Navigate to "+pageTitle+"?") + "\n\n"
	promptContent += descStyle.Render("Press Enter to open this page") + "\n\n"
	promptContent += keyStyle.Render("[Enter] Open  ")

	if s.notSwitchedMsg != "" && s.notSwitchedTimer > 0 {
		promptContent += skipKeyStyle.Render("[Esc] Skip")
	} else {
		promptContent += skipKeyStyle.Render("[Esc] Skip")
	}

	return promptStyle.Render(promptContent)
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

	if s.showPagePrompt {
		return s.renderPagePrompt()
	}

	if s.Width < 20 || s.Height < 10 {
		return "Terminal too small"
	}

	if s.currentSection >= len(s.boxes) {
		return ""
	}

	// Fixed 25% sidebar with minimum constraints
	sidebarWidth := int(float64(s.Width) * 0.25)
	// Minimum sidebar width: : don't exceed terminal15 chars, Maximum
	if sidebarWidth < 15 {
		sidebarWidth = 15
	}
	// Content width: remaining space minus borders and gap
	contentWidth := s.Width - sidebarWidth - 4 // -4 for both boxes borders (2+2)

	if contentWidth < 10 {
		contentWidth = 10
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(contentWidth).
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

	// Sidebar styles
	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(sidebarWidth).
		Height(s.getContentHeight() + 2).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	sidebarHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Bold(true).
		Underline(true)

	sidebarActiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	sidebarInactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669"))

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

	// Build sidebar content
	var sidebarLines []string
	sidebarLines = append(sidebarLines, sidebarHeaderStyle.Render("Sections"))
	sidebarLines = append(sidebarLines, "")

	// Calculate available width for section labels (account for borders, padding, and prefix)
	// sidebarWidth - 2 (borders) - 2 (padding) - 2 (prefix "→ " or "  ") = usable space
	maxLabelWidth := sidebarWidth - 6
	if maxLabelWidth < 1 {
		maxLabelWidth = 1
	}

	for i, title := range s.sectionTitles {
		sectionLabel := fmt.Sprintf("[%d] %s", i+1, title)
		// Truncate if exceeds available width
		if len(sectionLabel) > maxLabelWidth {
			sectionLabel = truncateString(sectionLabel, maxLabelWidth)
		}
		if i == s.currentSection {
			sidebarLines = append(sidebarLines, sidebarActiveStyle.Render("→ "+sectionLabel))
		} else {
			sidebarLines = append(sidebarLines, sidebarInactiveStyle.Render("  "+sectionLabel))
		}
	}

	// Join sidebar lines and render
	sidebarContent := strings.Join(sidebarLines, "\n")
	sidebarRendered := sidebarStyle.Render(sidebarContent)

	// Join content lines and render
	contentRendered := boxStyle.Render(strings.Join(contentLines, "\n"))

	// Combine content and sidebar horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, contentRendered, sidebarRendered)

	result := indicatorText + "\n" + mainContent + "\n" + controllerLine

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

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
