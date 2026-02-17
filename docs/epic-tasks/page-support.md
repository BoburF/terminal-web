# Page Switch Implementation Plan

## Current State Analysis

The terminal-web application currently supports:
- Single HTML page (`resume/index.html`) with multiple sections
- Section navigation using Tab/Shift+Tab keys
- Sidebar showing all sections with current section highlighted
- Section-based scrolling within each section

### Current Limitations
- Only one page can be displayed at a time
- No concept of multi-page navigation
- All content must be in a single HTML file
- No way to organize related content across multiple files

## Proposed Solution: Page Tree Navigation

Instead of traditional page switching, implement a **Page Tree** system where:
- Multiple HTML files act as separate "pages"
- Pages appear as special sections in the sidebar (tree structure)
- Navigating to a page section shows a confirmation prompt
- User can choose to enter the page or skip to next section
- Pages maintain their own section structure

### User Flow

```
┌─────────────────────────────────────────────────────────┐
│ Sections              │ Content Area                    │
│                       │                                 │
│ [1] Contact           │ ┌─────────────────────────┐     │
│ [2] About             │ │  About                  │     │
│ [3] Technical Skills  │ │  Backend Engineer...    │     │
│ [4] Experience        │ │                         │     │
│→[5] View Portfolio    │ │  Press Enter to open    │     │
│  [6] Education        │ │  this page              │     │
│  [7] Certifications   │ │                         │     │
│                       │ │  [Enter] Open Page      │     │
│                       │ │  [Esc]   Skip           │     │
│                       │ └─────────────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

When user navigates to "View Portfolio" (a page link):
1. Display confirmation prompt with page title
2. If Enter/Yes: Load the portfolio.html page
3. If Esc/No: Skip to next section, show "Not switched"

## Version Roadmap

### Version 1.0 - Page Discovery and Basic Navigation

#### Goals
- Detect and catalog all HTML files in resume directory
- Display pages as special sections in sidebar
- Implement page entry confirmation prompt
- Navigate to selected page on confirmation

#### Implementation Steps

**Step 1: Page Discovery System**
```go
// page_manager.go

type PageInfo struct {
    Filename    string
    Title       string
    Description string
    Order       int
}

func DiscoverPages(rootPath string) ([]PageInfo, error) {
    // Scan resume/ directory for *.html files
    // Parse each file to extract page-title from <body> or <head>
    // Return sorted list by order attribute or filename
}
```

**Step 2: Enhanced State Structure**
```go
// types.go additions

type Page struct {
    Info          PageInfo
    Boxes         []Box
    SectionTitles []string
    Loaded        bool
}

type State struct {
    // Existing fields
    Width, Height int
    boxes         []Box
    sectionTitles []string
    currentSection int
    sectionScrollOffset int
    
    // New page-related fields
    pages           []Page          // All discovered pages
    currentPageIdx  int             // Currently loaded page index
    pageHistory     []int           // Stack for back navigation
    showPagePrompt  bool            // Show confirmation dialog
    pendingPageIdx  int             // Page awaiting confirmation
    promptMessage   string          // Prompt text to display
}
```

**Step 3: Page Link Section Type**
```go
// HTML parsing extension
// Support new section type: <div section-type="page-link" page-target="portfolio.html">

func parseMain(doc *html.Node) ([]Box, []string) {
    // Existing logic for regular sections
    // Add detection for section-type="page-link"
    // Store target page filename in Box struct
}
```

**Step 4: Confirmation Prompt UI**
```go
func (s State) renderPagePrompt() string {
    // Render centered dialog box
    // Show page title and description
    // Display options: [Enter] Open Page  [Esc] Skip
}
```

**Step 5: Page Navigation Handlers**
```go
func (s State) handlePageNavigation() (State, tea.Cmd) {
    // When navigating to page-link section:
    // - Set showPagePrompt = true
    // - Store target page index in pendingPageIdx
    // - Display confirmation dialog
}

func (s State) confirmPageSwitch() (State, tea.Cmd) {
    // On Enter/Yes:
    // - Load target page
    // - Push current page to history
    // - Reset section index to 0
    // - Clear prompt
}

func (s State) cancelPageSwitch() (State, tea.Cmd) {
    // On Esc/No:
    // - Show "Not switched" message briefly
    // - Move to next section or stay on current
    // - Clear prompt
}
```

#### Checking Functions

```go
// page_manager_test.go

func TestDiscoverPages(t *testing.T) {
    pages, err := DiscoverPages("./test-resume/")
    require.NoError(t, err)
    assert.GreaterOrEqual(t, len(pages), 1)
    assert.Equal(t, "index.html", pages[0].Filename)
}

func TestPageInfoExtraction(t *testing.T) {
    html := `<body page-title="Portfolio" page-desc="My work showcase">`
    info := ExtractPageInfo(html)
    assert.Equal(t, "Portfolio", info.Title)
    assert.Equal(t, "My work showcase", info.Description)
}

func TestPageLinkParsing(t *testing.T) {
    // Test that page-link sections are correctly identified
}
```

#### Success Criteria
- [ ] All HTML files in resume/ are discovered
- [ ] Pages appear in sidebar with distinctive styling
- [ ] Confirmation prompt displays correctly
- [ ] Enter opens the page, Esc skips
- [ ] "Not switched" message appears briefly on skip

---

### Version 1.1 - Page Breadcrumbs and History

#### Goals
- Show breadcrumb trail: Main > Portfolio > Contact
- Implement back navigation with history stack
- Visual indication of current location in page tree

#### Implementation Steps

**Step 1: Breadcrumb Rendering**
```go
func (s State) renderBreadcrumbs() string {
    // Build trail: Page1 > Page2 > Current
    // Show in header area
}
```

**Step 2: History Stack**
```go
func (s *State) pushHistory(pageIdx int) {
    s.pageHistory = append(s.pageHistory, pageIdx)
}

func (s *State) popHistory() (int, bool) {
    if len(s.pageHistory) == 0 {
        return 0, false
    }
    idx := s.pageHistory[len(s.pageHistory)-1]
    s.pageHistory = s.pageHistory[:len(s.pageHistory)-1]
    return idx, true
}
```

**Step 3: Back Navigation**
- Add `b` key binding for "back"
- Pop from history and load previous page
- Restore previous section and scroll position

**Step 4: Enhanced Prompt**
```
┌─────────────────────────────────────┐
│ Navigate to Portfolio?              │
│                                     │
│ My work showcase                    │
│                                     │
│ [Enter] Open    [Esc] Skip    [?]   │
│                                     │
└─────────────────────────────────────┘
```

#### Checking Functions

```go
func TestHistoryStack(t *testing.T) {
    s := State{currentPageIdx: 0}
    s.pushHistory(0)
    s.navigateToPage(2)
    s.pushHistory(2)
    
    prev, ok := s.popHistory()
    assert.True(t, ok)
    assert.Equal(t, 0, prev)
}

func TestBreadcrumbTrail(t *testing.T) {
    s := createMultiPageState()
    trail := s.getBreadcrumbTrail()
    assert.Contains(t, trail, "index")
    assert.Contains(t, trail, "portfolio")
}
```

#### Success Criteria
- [ ] Breadcrumb shows current page path
- [ ] Back navigation works correctly
- [ ] History maintains correct order
- [ ] Position restored when going back

---

### Version 1.2 - Page State Persistence

#### Goals
- Remember scroll position and section when leaving a page
- Restore exact state when returning via history
- Optional: Persist across application restarts

#### Implementation Steps

**Step 1: Page State Storage**
```go
type PagePosition struct {
    SectionIndex   int
    ScrollOffset   int
    LastVisited    time.Time
}

// Map page index to saved position
pagePositions map[int]PagePosition
```

**Step 2: Save on Exit**
```go
func (s *State) saveCurrentPosition() {
    s.pagePositions[s.currentPageIdx] = PagePosition{
        SectionIndex: s.currentSection,
        ScrollOffset: s.sectionScrollOffset,
        LastVisited:  time.Now(),
    }
}
```

**Step 3: Restore on Enter**
```go
func (s *State) restorePosition(pageIdx int) {
    if pos, exists := s.pagePositions[pageIdx]; exists {
        s.currentSection = pos.SectionIndex
        s.sectionScrollOffset = pos.ScrollOffset
    } else {
        s.currentSection = 0
        s.sectionScrollOffset = 0
    }
}
```

**Step 4: Visual Indicator**
- Show "visited" indicator on page links
- Optional: Show last visited time

#### Checking Functions

```go
func TestPositionSaveRestore(t *testing.T) {
    s := createTestState()
    s.currentSection = 3
    s.sectionScrollOffset = 10
    s.saveCurrentPosition()
    
    s.currentSection = 0
    s.sectionScrollOffset = 0
    s.restorePosition(s.currentPageIdx)
    
    assert.Equal(t, 3, s.currentSection)
    assert.Equal(t, 10, s.sectionScrollOffset)
}
```

#### Success Criteria
- [ ] Position saved when leaving page
- [ ] Position restored when returning
- [ ] Works with back navigation
- [ ] Visual indicator for visited pages

---

### Version 1.3 - Quick Navigation and Search

#### Goals
- Quick jump to any page with `/` or `Space`
- Fuzzy search across page titles
- List all pages with `l` key

#### Implementation Steps

**Step 1: Page List View**
```go
func (s State) renderPageList() string {
    // Full-screen list of all pages
    // Numbered for quick selection
    // Show description and visited status
}
```

**Step 2: Fuzzy Search**
```go
func fuzzySearchPages(pages []Page, query string) []Page {
    // Simple fuzzy matching on title and description
    // Return sorted by relevance
}
```

**Step 3: Quick Jump Input**
```go
// Add text input for search
// Real-time filtering as user types
// Enter to select highlighted page
```

#### Checking Functions

```go
func TestFuzzySearch(t *testing.T) {
    pages := []Page{
        {Info: PageInfo{Title: "Portfolio"}},
        {Info: PageInfo{Title: "About Me"}},
        {Info: PageInfo{Title: "Contact"}},
    }
    
    results := fuzzySearchPages(pages, "port")
    assert.Equal(t, "Portfolio", results[0].Info.Title)
}
```

#### Success Criteria
- [ ] Page list displays all pages
- [ ] Fuzzy search returns relevant results
- [ ] Quick navigation with number keys
- [ ] Search works with partial matches

---

### Version 1.4 - Advanced Tree Navigation

#### Goals
- Support nested page hierarchies
- Expand/collapse page groups
- Visual tree structure in sidebar

#### Implementation Steps

**Step 1: Hierarchical Structure**
```go
type PageNode struct {
    Page     PageInfo
    Children []PageNode
    Expanded bool
    Level    int
}
```

**Step 2: Tree Rendering**
```go
func renderTree(nodes []PageNode, currentIdx int) string {
    // Render with indentation based on level
    // Show expand/collapse indicators
    // ├─ Portfolio
    // │  ├─ Web Projects
    // │  └─ Mobile Apps
    // └─ About
}
```

**Step 3: Tree Controls**
- `→` to expand group
- `←` to collapse group
- `Enter` to open leaf page

#### Checking Functions

```go
func TestTreeNavigation(t *testing.T) {
    tree := buildPageTree(testFiles)
    assert.Equal(t, 2, len(tree.Children))
    
    // Test expand/collapse
    tree.Children[0].Expanded = true
    visible := countVisibleNodes(tree)
    assert.Greater(t, visible, 2)
}
```

#### Success Criteria
- [ ] Tree structure displays correctly
- [ ] Expand/collapse works
- [ ] Navigation respects hierarchy
- [ ] Visual tree indicators render properly

---

## HTML Specification

### Page Definition

Create new HTML files in `resume/` directory:

```html
<!-- resume/portfolio.html -->
<head>
    <link rel="script" type="lua" href="./portfolio.lua" />
</head>
<body page-title="Portfolio" 
      page-description="My development work showcase"
      page-order="2">
    
    <div class="main">
        <div section-title="Web Projects">
            <h1>Web Development</h1>
            <p>E-commerce platform...</p>
        </div>
        
        <div section-title="Mobile Apps">
            <h1>Mobile Applications</h1>
            <p>iOS and Android apps...</p>
        </div>
    </div>
    
    <div class="controllers">
        <button type="exit" bind="q">Exit</button>
        <button type="back" bind="b">Back</button>
    </div>
</body>
```

### Page Link Sections

In main `index.html`, add links to other pages:

```html
<!-- resume/index.html -->
<div section-title="Experience">
    <h1>Work Experience</h1>
    <p>Backend Engineer...</p>
</div>

<!-- Page link section - appears in sidebar with special styling -->
<div section-type="page-link" 
     page-target="portfolio.html"
     section-title="View Portfolio">
    <h1>Portfolio</h1>
    <p>Press Enter to view my projects</p>
</div>

<div section-title="Education">
    <h1>Education</h1>
    <p>Computer Science...</p>
</div>
```

### Supported Attributes

**Body Attributes:**
- `page-title` - Display name in sidebar and breadcrumbs
- `page-description` - Shown in confirmation prompt
- `page-order` - Sorting order (optional)
- `page-category` - For grouping in tree view

**Section Attributes:**
- `section-type="page-link"` - Marks section as page navigation
- `page-target` - Target HTML filename
- `section-title` - Label in sidebar

---

## Control Reference

### Global Controls

| Key | Action | Context |
|-----|--------|---------|
| `Tab` | Next section | Always |
| `Shift+Tab` | Previous section | Always |
| `Enter` | Confirm/Select | In prompt |
| `Esc` | Cancel/Skip | In prompt |
| `b` | Go back | After navigation |
| `q` / `Ctrl+C` | Exit | Always |
| `j` / `↓` | Scroll down | In section |
| `k` / `↑` | Scroll up | In section |

### Page Prompt Controls

| Key | Action |
|-----|--------|
| `Enter` / `y` | Open selected page |
| `Esc` / `n` | Skip page, show "Not switched" |
| `?` | Show help |

### Navigation Controls (v1.1+)

| Key | Action | Version |
|-----|--------|---------|
| `b` | Back to previous page | 1.1 |
| `l` | List all pages | 1.3 |
| `/` or `Space` | Quick search | 1.3 |
| number | Jump to page N | 1.3 |

### Tree Controls (v1.4)

| Key | Action |
|-----|--------|
| `→` | Expand group |
| `←` | Collapse group |
| `Enter` | Open page or toggle group |

---

## UI Flow Diagrams

### Page Entry Flow

```
User navigates to page-link section
         ↓
Display confirmation prompt
         ↓
    ┌────┴────┐
    ↓         ↓
  Enter      Esc
    ↓         ↓
Load page   Show "Not switched"
    ↓         ↓
Render     Move to next section
new page
```

### History Navigation Flow

```
User opens Page A
      ↓
Navigate to Page B
      ↓
Press 'b' for back
      ↓
Restore Page A
      ↓
Restore section
and scroll position
```

### State Machine

```
┌─────────┐     Tab      ┌──────────────┐
│ VIEWING ├─────────────→│ PAGE_PROMPT  │
│ SECTION │               │ (Confirm?)   │
└────┬────┘               └──────┬───────┘
     ↑                            │
     │        ┌──────────┐        │
     └────────┤ CANCELLED├←───────┘ Esc
              │ (skip)   │
              └────┬─────┘
                   │ Enter
              ┌────┴─────┐
              │ PAGE_VIEW│
              │ (new pg) │
              └────┬─────┘
                   │ 'b'
              ┌────┴─────┐
              │  BACK    │
              │(restore) │
              └──────────┘
```

---

## Implementation Files

### New Files

1. **internal/page/manager.go**
   - Page discovery and loading
   - Cache management

2. **internal/page/types.go**
   - Page and PageInfo structs
   - PagePosition for state

3. **internal/ui/prompt.go**
   - Confirmation prompt rendering
   - Prompt state management

4. **internal/ui/breadcrumbs.go**
   - Breadcrumb trail generation
   - History management

5. **internal/navigation/history.go**
   - History stack operations
   - Position save/restore

### Modified Files

1. **main.go**
   - Multi-page initialization
   - Page discovery on startup

2. **buble.go**
   - Add page-related state fields
   - Handle page prompt state
   - Render breadcrumbs

3. **tui.go**
   - Parse page-link sections
   - Handle page navigation events
   - Update View() for prompts

4. **resume/index.html**
   - Add page-link sections
   - Add back button controller

---

## Testing Checklist

### Unit Tests

- [ ] Page discovery finds all HTML files
- [ ] Page info extraction from attributes
- [ ] Page link parsing identifies targets
- [ ] History push/pop operations
- [ ] Position save/restore
- [ ] Fuzzy search ranking

### Integration Tests

- [ ] Full navigation flow between pages
- [ ] History navigation after multiple jumps
- [ ] Position restoration accuracy
- [ ] Prompt display and dismissal
- [ ] "Not switched" message display

### Manual Testing

- [ ] Tab navigation reaches page links
- [ ] Prompt renders correctly
- [ ] Enter opens page
- [ ] Esc shows skip message
- [ ] Back returns to previous page
- [ ] Breadcrumbs update correctly
- [ ] Sidebar updates for new page
- [ ] All sections accessible on new page

### Edge Cases

- [ ] Page file not found (graceful error)
- [ ] Empty page (no sections)
- [ ] Circular page references
- [ ] Very long page titles
- [ ] Rapid navigation (key spam)
- [ ] Terminal resize during prompt

---

## Migration Guide

### Existing Single-Page Setup

1. **No changes required** - index.html continues to work
2. Pages are optional enhancement
3. Gradually add page-link sections

### Adding First Page

1. Create `resume/portfolio.html`:
```html
<body page-title="Portfolio">
    <div class="main">
        <div section-title="Projects">
            <h1>My Projects</h1>
            <p>Project details...</p>
        </div>
    </div>
    <div class="controllers">
        <button type="exit" bind="q">Exit</button>
    </div>
</body>
```

2. Add link in `index.html`:
```html
<div section-type="page-link" 
     page-target="portfolio.html"
     section-title="View Portfolio">
    <h1>Portfolio</h1>
    <p>See my work examples</p>
</div>
```

3. Restart application
4. Navigate to new section
5. Press Enter to open portfolio

---

## Future Enhancements (Post-v1.4)

- **Page transitions**: Animated transitions between pages
- **Preloading**: Background loading of linked pages
- **Bookmarks**: User-defined favorite pages
- **Session sharing**: Export/import navigation history
- **Page metadata**: Author, date, tags for filtering
- **Conditional links**: Show/hide based on Lua conditions

---

## Performance Considerations

### Page Loading
- Parse pages on demand (lazy loading)
- Cache parsed page structures
- LRU cache for recently visited pages

### Memory Management
- Unload inactive pages after threshold
- Limit history stack size
- Clear cache on low memory

### Rendering
- Only render visible pages
- Virtual scrolling for large page lists
- Debounce rapid navigation

---

## Summary

This implementation plan provides a progressive approach to adding multi-page support while maintaining the existing single-page experience. The tree-based navigation with confirmation prompts provides intuitive page discovery without overwhelming users.

Key innovations:
- **Page sections**: Pages appear as sections in the sidebar tree
- **Confirmation prompts**: Explicit user consent before navigation
- **History preservation**: Full state restoration when returning
- **Progressive enhancement**: Each version adds features without breaking previous

The system transforms the TUI from a single-document viewer into a lightweight hypertext navigation system while maintaining terminal-friendly keyboard controls.
