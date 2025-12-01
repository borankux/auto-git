package ui

import (
	"fmt"
	"io"
	"strings"

	"auto-git/internal/provider"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().MarginTop(2).MarginLeft(4)
)

type Model struct {
	list     list.Model
	choice   string
	quitting bool
}

type item struct {
	title, desc string
}

func (i item) FilterValue() string { return i.title }

type modelSelectionModel struct {
	list   list.Model
	choice string
}

func (m modelSelectionModel) Init() tea.Cmd {
	return nil
}

func (m modelSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.title
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelSelectionModel) View() string {
	if m.choice != "" {
		return ""
	}
	return "\n" + m.list.View()
}

func SelectModel(models []provider.Model, defaultModel string) (string, error) {
	items := make([]list.Item, len(models))
	selectedIndex := 0

	for i, m := range models {
		desc := ""
		if m.Size > 0 {
			desc = fmt.Sprintf("Size: %d bytes", m.Size)
		} else if m.ModifiedAt != "" {
			desc = fmt.Sprintf("Modified: %s", m.ModifiedAt)
		}
		items[i] = item{title: m.Name, desc: desc}
		if m.Name == defaultModel {
			selectedIndex = i
		}
	}

	l := list.New(items, itemDelegate{}, 80, 20)
	l.Title = "Select Model"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle()
	l.Styles.HelpStyle = helpStyle
	l.Select(selectedIndex)

	m := modelSelectionModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run UI: %w", err)
	}

	if m, ok := finalModel.(modelSelectionModel); ok {
		if m.choice != "" {
			return m.choice, nil
		}
	}

	if len(models) > 0 {
		return models[selectedIndex].Name, nil
	}

	return defaultModel, nil
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type messageEditModel struct {
	textInput textinput.Model
	message   string
	done      bool
}

func (m messageEditModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m messageEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			m.message = ""
			return m, tea.Quit

		case "enter":
			m.done = true
			m.message = m.textInput.Value()
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m messageEditModel) View() string {
	return fmt.Sprintf(
		"\nEdit commit message:\n\n%s\n\n%s",
		m.textInput.View(),
		"(enter to confirm, esc to cancel)",
	) + "\n"
}

func EditCommitMessage(initialMessage string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = "Enter commit message..."
	ti.SetValue(initialMessage)
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 80

	m := messageEditModel{
		textInput: ti,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run UI: %w", err)
	}

	if m, ok := finalModel.(messageEditModel); ok {
		if m.done {
			return m.message, nil
		}
	}

	return "", fmt.Errorf("message editing cancelled")
}

type progressModel struct {
	message string
	done    bool
}

func (m progressModel) Init() tea.Cmd {
	return nil
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case progressDoneMsg:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m progressModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("\n%s\n", m.message)
}

type progressDoneMsg struct{}

func ShowProgress(message string) {
	m := progressModel{message: message}
	p := tea.NewProgram(m)
	p.Run()
}

