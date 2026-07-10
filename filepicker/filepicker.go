// Package filepicker provides a file picker component for Bubble Tea
// applications.
package filepicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/sahilm/fuzzy"
)

type SortBy int

const (
	SortByName SortBy = iota
	SortByDatetime
)

func (s SortBy) String() string {
	switch s {
	case SortByName:
		return "Name"
	case SortByDatetime:
		return "Datetime"
	default:
		return "Unknown"
	}
}

type SortOrder int

const (
	SortAscending SortOrder = iota
	SortDescending
)

func (o SortOrder) String() string {
	switch o {
	case SortAscending:
		return "asc"
	case SortDescending:
		return "desc"
	default:
		return "unknown"
	}
}

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

// New returns a new filepicker model with default styling and key bindings.
func New() Model {
	return Model{
		id:               nextID(),
		CurrentDirectory: ".",
		Cursor:           ">",
		AllowedTypes:     []string{},
		selected:         0,
		ShowPermissions:  true,
		ShowSize:         true,
		ShowHidden:       true,
		DirAllowed:       false,
		FileAllowed:      true,
		AutoHeight:       true,
		Height:           0,
		max:              0,
		min:              0,
		selectedStack:    newStack(),
		minStack:         newStack(),
		maxStack:         newStack(),
		KeyMap:           DefaultKeyMap(),
		Styles:           DefaultStyles(),
		SortBy:           SortByName,
		SortOrder:        SortAscending,
		Query:            "",
	}
}

type errorMsg struct {
	err error
}

type readDirMsg struct {
	id      int
	entries []os.DirEntry
}

const (
	marginBottom  = 5
	fileSizeWidth = 7
	paddingLeft   = 2
)

// KeyMap defines key bindings for each user action.
type KeyMap struct {
	GoToTop   key.Binding
	GoToLast  key.Binding
	Down      key.Binding
	Up        key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	Back      key.Binding
	Open      key.Binding
	Select    key.Binding
	SortType  key.Binding
	SortOrder key.Binding
}

// DefaultKeyMap defines the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		GoToTop:   key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first")),
		GoToLast:  key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last")),
		Down:      key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "down")),
		Up:        key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "up")),
		PageUp:    key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown:  key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
		Back:      key.NewBinding(key.WithKeys("left"), key.WithHelp("left", "parent")),
		Open:      key.NewBinding(key.WithKeys("right", "enter"), key.WithHelp("enter", "open")),
		Select:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		SortType:  key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "toggle sort")),
		SortOrder: key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "toggle sort order")),
	}
}

// Styles defines the possible customizations for styles in the file picker.
type Styles struct {
	DisabledCursor   lipgloss.Style
	Cursor           lipgloss.Style
	Symlink          lipgloss.Style
	Directory        lipgloss.Style
	File             lipgloss.Style
	DisabledFile     lipgloss.Style
	Permission       lipgloss.Style
	Selected         lipgloss.Style
	DisabledSelected lipgloss.Style
	FileSize         lipgloss.Style
	EmptyDirectory   lipgloss.Style
}

// DefaultStyles defines the default styling for the file picker.
func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer defines the default styling for the file picker,
// with a given Lip Gloss renderer.
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		DisabledCursor:   r.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:           r.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:          r.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:        r.NewStyle().Foreground(lipgloss.Color("99")),
		File:             r.NewStyle(),
		DisabledFile:     r.NewStyle().Foreground(lipgloss.Color("243")),
		DisabledSelected: r.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:       r.NewStyle().Foreground(lipgloss.Color("244")),
		Selected:         r.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		FileSize:         r.NewStyle().Foreground(lipgloss.Color("240")).Width(fileSizeWidth).Align(lipgloss.Right),
		EmptyDirectory:   r.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(paddingLeft).SetString("Bummer. No Files Found."),
	}
}

// Model represents a file picker.
type Model struct {
	id int

	// Path is the path which the user has selected with the file picker.
	Path string

	// CurrentDirectory is the directory that the user is currently in.
	CurrentDirectory string

	// AllowedTypes specifies which file types the user may select.
	// If empty the user may select any file.
	AllowedTypes []string

	KeyMap          KeyMap
	allFiles        []os.DirEntry
	files           []os.DirEntry
	ShowPermissions bool
	ShowSize        bool
	ShowHidden      bool
	DirAllowed      bool
	FileAllowed     bool

	FileSelected  string
	selected      int
	selectedStack stack

	min      int
	max      int
	maxStack stack
	minStack stack

	// Height of the picker.
	//
	// Deprecated: use [Model.SetHeight] instead.
	Height     int
	AutoHeight bool

	Cursor    string
	Styles    Styles
	SortBy    SortBy
	SortOrder SortOrder
	Query     string
}

type stack struct {
	Push   func(int)
	Pop    func() int
	Length func() int
}

func newStack() stack {
	slice := make([]int, 0)
	return stack{
		Push: func(i int) {
			slice = append(slice, i)
		},
		Pop: func() int {
			res := slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			return res
		},
		Length: func() int {
			return len(slice)
		},
	}
}

func (m *Model) pushView(selected, minimum, maximum int) {
	m.selectedStack.Push(selected)
	m.minStack.Push(minimum)
	m.maxStack.Push(maximum)
}

func (m *Model) popView() (int, int, int) {
	return m.selectedStack.Pop(), m.minStack.Pop(), m.maxStack.Pop()
}

func sortEntries(entries []os.DirEntry, sortBy SortBy, sortOrder SortOrder) {
	if sortBy == SortByDatetime {
		type datedEntry struct {
			entry   os.DirEntry
			modTime int64
		}
		dated := make([]datedEntry, len(entries))
		for i, entry := range entries {
			dated[i].entry = entry
			if info, err := entry.Info(); err == nil {
				dated[i].modTime = info.ModTime().UnixNano()
			}
		}
		sort.SliceStable(dated, func(i, j int) bool {
			if dated[i].entry.IsDir() != dated[j].entry.IsDir() {
				return dated[i].entry.IsDir()
			}
			if dated[i].modTime == dated[j].modTime {
				return strings.ToLower(dated[i].entry.Name()) < strings.ToLower(dated[j].entry.Name())
			}
			if sortOrder == SortAscending {
				return dated[i].modTime < dated[j].modTime
			}
			return dated[i].modTime > dated[j].modTime
		})
		for i := range dated {
			entries[i] = dated[i].entry
		}
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		if sortOrder == SortAscending {
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		}
		return strings.ToLower(entries[i].Name()) > strings.ToLower(entries[j].Name())
	})
}

func (m *Model) sortFiles() {
	sortEntries(m.allFiles, m.SortBy, m.SortOrder)
	m.applyFilter()
}

func (m *Model) resetViewport() {
	m.selected = 0
	m.min = 0
	m.max = max(m.Height-1, 0)
	if len(m.files) > 0 && m.max >= len(m.files) {
		m.max = len(m.files) - 1
	}
}

func (m *Model) applyFilter() {
	query := strings.TrimSpace(m.Query)
	if query == "" {
		m.files = append(m.files[:0], m.allFiles...)
		m.resetViewport()
		return
	}

	names := make([]string, len(m.allFiles))
	for i, entry := range m.allFiles {
		names[i] = entry.Name()
	}
	matches := fuzzy.Find(query, names)
	filtered := make([]os.DirEntry, 0, len(matches))
	for _, match := range matches {
		filtered = append(filtered, m.allFiles[match.Index])
	}
	m.files = filtered
	m.resetViewport()

}

// ResetFilter clears the type-ahead query while keeping the current directory.
func (m *Model) ResetFilter() {
	m.Query = ""
	m.applyFilter()
}

// MatchCount returns the visible and total entry counts for the current directory.
func (m Model) MatchCount() (int, int) {
	return len(m.files), len(m.allFiles)
}

func (m Model) readDir(path string, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg{err}
		}

		sortEntries(dirEntries, m.SortBy, m.SortOrder)

		if showHidden {
			return readDirMsg{id: m.id, entries: dirEntries}
		}

		var sanitizedDirEntries []os.DirEntry
		for _, dirEntry := range dirEntries {
			isHidden, _ := IsHidden(dirEntry.Name())
			if isHidden {
				continue
			}
			sanitizedDirEntries = append(sanitizedDirEntries, dirEntry)
		}
		return readDirMsg{id: m.id, entries: sanitizedDirEntries}
	}
}

// Init initializes the file picker model.
func (m Model) Init() tea.Cmd {
	return m.readDir(m.CurrentDirectory, m.ShowHidden)
}

// SetHeight sets the height of the filepicker.
func (m *Model) SetHeight(height int) {
	m.Height = height
	if m.max > m.Height-1 {
		m.max = m.min + m.Height - 1
	}
}

// Update handles user interactions within the file picker model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case readDirMsg:
		if msg.id != m.id {
			break
		}
		m.allFiles = msg.entries
		m.applyFilter()
	case tea.WindowSizeMsg:
		if m.AutoHeight {
			m.Height = msg.Height - marginBottom
		}
		m.max = m.Height - 1
	case tea.KeyMsg:
		m.Path = ""
		switch {
		case key.Matches(msg, m.KeyMap.SortType):
			if m.SortBy == SortByName {
				m.SortBy = SortByDatetime
			} else {
				m.SortBy = SortByName
			}
			m.sortFiles()
			return m, nil

		case key.Matches(msg, m.KeyMap.SortOrder):
			if m.SortOrder == SortAscending {
				m.SortOrder = SortDescending
			} else {
				m.SortOrder = SortAscending
			}
			m.sortFiles()
			return m, nil

		case key.Matches(msg, m.KeyMap.GoToTop):
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
		case key.Matches(msg, m.KeyMap.GoToLast):
			if len(m.files) == 0 {
				break
			}
			m.selected = len(m.files) - 1
			m.min = max(len(m.files)-m.Height, 0)
			m.max = len(m.files) - 1
		case key.Matches(msg, m.KeyMap.Down):
			if len(m.files) == 0 {
				break
			}
			m.selected++
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			if m.selected > m.max {
				m.min++
				m.max++
			}
		case key.Matches(msg, m.KeyMap.Up):
			if len(m.files) == 0 {
				break
			}
			m.selected--
			if m.selected < 0 {
				m.selected = 0
			}
			if m.selected < m.min {
				m.min--
				m.max--
			}
		case key.Matches(msg, m.KeyMap.PageDown):
			if len(m.files) == 0 {
				break
			}
			m.selected += m.Height
			if m.selected >= len(m.files) {
				m.selected = len(m.files) - 1
			}
			m.min += m.Height
			m.max += m.Height

			if m.max >= len(m.files) {
				m.max = len(m.files) - 1
				m.min = max(m.max-m.Height+1, 0)
			}
		case key.Matches(msg, m.KeyMap.PageUp):
			if len(m.files) == 0 {
				break
			}
			m.selected -= m.Height
			if m.selected < 0 {
				m.selected = 0
			}
			m.min -= m.Height
			m.max -= m.Height

			if m.min < 0 {
				m.min = 0
				m.max = m.min + m.Height
			}
		case key.Matches(msg, m.KeyMap.Back):
			m.CurrentDirectory = filepath.Dir(m.CurrentDirectory)
			m.Query = ""
			if m.selectedStack.Length() > 0 {
				m.selected, m.min, m.max = m.popView()
			} else {
				m.selected = 0
				m.min = 0
				m.max = m.Height - 1
			}
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case key.Matches(msg, m.KeyMap.Open):
			if len(m.files) == 0 {
				break
			}

			f := m.files[m.selected]
			info, err := f.Info()
			if err != nil {
				break
			}
			isSymlink := info.Mode()&os.ModeSymlink != 0
			isDir := f.IsDir()

			if isSymlink {
				symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
				info, err := os.Stat(symlinkPath)
				if err != nil {
					break
				}
				if info.IsDir() {
					isDir = true
				}
			}

			if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
				if key.Matches(msg, m.KeyMap.Select) {
					// Select the current path as the selection
					m.Path = filepath.Join(m.CurrentDirectory, f.Name())
				}
			}

			if !isDir {
				break
			}

			m.CurrentDirectory = filepath.Join(m.CurrentDirectory, f.Name())
			m.Query = ""
			m.pushView(m.selected, m.min, m.max)
			m.selected = 0
			m.min = 0
			m.max = m.Height - 1
			return m, m.readDir(m.CurrentDirectory, m.ShowHidden)
		case msg.String() == "backspace":
			queryRunes := []rune(m.Query)
			if len(queryRunes) > 0 {
				m.Query = string(queryRunes[:len(queryRunes)-1])
				m.applyFilter()
			}
		case msg.String() == "ctrl+u":
			m.Query = ""
			m.applyFilter()
		case msg.Type == tea.KeyRunes && !msg.Alt:
			m.Query += string(msg.Runes)
			m.applyFilter()
		}
	}
	return m, nil
}

// View returns the view of the file picker.
func (m Model) View() string {
	if len(m.files) == 0 {
		return m.Styles.EmptyDirectory.Height(m.Height).MaxHeight(m.Height).String()
	}
	var s strings.Builder

	for i, f := range m.files {
		if i < m.min || i > m.max {
			continue
		}

		var symlinkPath string
		info, err := f.Info()
		if err != nil {
			continue
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		size := "-"
		if info.Mode().IsRegular() && info.Size() >= 0 {
			size = strings.Replace(humanize.Bytes(uint64(info.Size())), " ", "", 1) // #nosec G115 -- non-negative size is checked immediately above.
		}
		name := f.Name()

		if isSymlink {
			symlinkPath, _ = filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, name))
		}

		disabled := !m.canSelect(name) && !f.IsDir()

		if m.selected == i { //nolint:nestif
			selected := ""
			if m.ShowPermissions {
				selected += " " + info.Mode().String()
			}
			if m.ShowSize {
				selected += fmt.Sprintf("%"+strconv.Itoa(m.Styles.FileSize.GetWidth())+"s", size)
			}
			selected += " " + name
			if isSymlink {
				selected += " → " + symlinkPath
			}
			if disabled {
				s.WriteString(m.Styles.DisabledSelected.Render(m.Cursor) + m.Styles.DisabledSelected.Render(selected))
			} else {
				s.WriteString(m.Styles.Cursor.Render(m.Cursor) + m.Styles.Selected.Render(selected))
			}
			s.WriteRune('\n')
			continue
		}

		style := m.Styles.File
		if f.IsDir() {
			style = m.Styles.Directory
		} else if isSymlink {
			style = m.Styles.Symlink
		} else if disabled {
			style = m.Styles.DisabledFile
		}

		fileName := style.Render(name)
		s.WriteString(m.Styles.Cursor.Render(" "))
		if isSymlink {
			fileName += " → " + symlinkPath
		}
		if m.ShowPermissions {
			s.WriteString(" " + m.Styles.Permission.Render(info.Mode().String()))
		}
		if m.ShowSize {
			s.WriteString(m.Styles.FileSize.Render(size))
		}
		s.WriteString(" " + fileName)
		s.WriteRune('\n')
	}

	for i := lipgloss.Height(s.String()); i < m.Height; i++ {
		s.WriteRune('\n')
	}

	return s.String()
}

// DidSelectFile returns whether a user has selected a file (on this msg).
func (m Model) DidSelectFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && m.canSelect(path) {
		return true, path
	}
	return false, ""
}

// DidSelectDisabledFile returns whether a user tried to select a disabled file
// (on this msg). This is necessary only if you would like to warn the user that
// they tried to select a disabled file.
func (m Model) DidSelectDisabledFile(msg tea.Msg) (bool, string) {
	didSelect, path := m.didSelectFile(msg)
	if didSelect && !m.canSelect(path) {
		return true, path
	}
	return false, ""
}

func (m Model) didSelectFile(msg tea.Msg) (bool, string) {
	if len(m.files) == 0 || m.Path == "" {
		return false, ""
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If the msg does not match the Select keymap then this could not have been a selection.
		if !key.Matches(msg, m.KeyMap.Select) {
			return false, ""
		}

		// The key press was a selection, let's confirm whether the current file could
		// be selected or used for navigating deeper into the stack.
		f := m.files[m.selected]
		info, err := f.Info()
		if err != nil {
			return false, ""
		}
		isSymlink := info.Mode()&os.ModeSymlink != 0
		isDir := f.IsDir()

		if isSymlink {
			symlinkPath, _ := filepath.EvalSymlinks(filepath.Join(m.CurrentDirectory, f.Name()))
			info, err := os.Stat(symlinkPath)
			if err != nil {
				break
			}
			if info.IsDir() {
				isDir = true
			}
		}

		if (!isDir && m.FileAllowed) || (isDir && m.DirAllowed) {
			return true, m.Path
		}

		// If the msg was not a KeyMsg, then the file could not have been selected this iteration.
		// Only a KeyMsg can select a file.
	default:
		return false, ""
	}
	return false, ""
}

func (m Model) canSelect(file string) bool {
	if len(m.AllowedTypes) <= 0 {
		return true
	}

	for _, ext := range m.AllowedTypes {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}
