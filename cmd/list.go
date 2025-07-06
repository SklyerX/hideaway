package cmd

import (
	"fmt"
	"hideaway/utils"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#F25D94")).
				Padding(0, 1)

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	altRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2A2A2A")).
			Padding(0, 1)

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Margin(1, 0)
)

type tableModel struct {
	data     []map[string]interface{}
	headers  []string
	cursor   int
	viewport struct {
		start int
		end   int
	}
	width       int
	height      int
	message     string
	isConfirmed bool
}

type clearMessageMsg struct{}

func newTableModel(jsonData []map[string]interface{}) tableModel {
	m := tableModel{
		data: jsonData,
	}

	// Extract headers from the first object
	if len(jsonData) > 0 {
		for key := range jsonData[0] {
			m.headers = append(m.headers, key)
		}
	}

	// Initialize viewport with default values
	m.viewport.start = 0
	m.viewport.end = len(jsonData) // Will be properly adjusted in first Update

	return m
}

func (m tableModel) Init() tea.Cmd {
	return nil
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Initialize viewport properly when we get size info
		visibleRows := m.height - 6 // Account for header, borders, and help text
		if visibleRows > 0 {
			m.viewport.end = min(len(m.data), visibleRows)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.data)-1 {
				m.cursor++
				// Auto-scroll viewport
				if m.cursor >= m.viewport.end {
					m.viewport.start++
					m.viewport.end++
				}
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				// Auto-scroll viewport
				if m.cursor < m.viewport.start {
					m.viewport.start--
					m.viewport.end--
				}
			}

		case "g":
			m.cursor = 0
			m.viewport.start = 0
			m.viewport.end = min(len(m.data), m.height-6)

		case "G":
			m.cursor = len(m.data) - 1
			if len(m.data) > m.height-6 {
				m.viewport.start = len(m.data) - (m.height - 6)
				m.viewport.end = len(m.data)
			}
		case "r":
			index := m.cursor
			dataFiles := m.data

			obj := dataFiles[index]

			id, ok := obj["Id"]
			if ok {
				if idStr, ok := id.(string); ok {
					utils.RetrieveFile(idStr, authenticatedPassword)
				} else {
					color.Red("Invalid ID type, expected string")
				}
			}

		case "d":
			m.message = "Press 'y' to confirm delete, 'n' to cancel"
			m.isConfirmed = true

		case "y":
			if !m.isConfirmed {
				return m, nil
			}

			m.message = "File deleted successfully!"
			m.isConfirmed = false
			// Do actual deletion here

			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearMessageMsg{}
			})

		case "n":
			if m.isConfirmed {
				m.message = "Delete cancelled"
				m.isConfirmed = false

				return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return clearMessageMsg{}
				})
			}

		}
	case clearMessageMsg:
		m.message = ""
		return m, nil
	}

	if m.viewport.end == 0 && m.height == 0 {
		m.viewport.end = min(len(m.data), 10)
	} else if m.height > 0 {
		visibleRows := m.height - 6
		if visibleRows > 0 && m.viewport.end == 0 {
			m.viewport.end = min(len(m.data), visibleRows)
		}
	}

	return m, nil
}

func (m tableModel) View() string {
	if len(m.data) == 0 {
		return "No data to display!"
	}

	availableWidth := m.width - 6
	if availableWidth < 40 {
		availableWidth = 40
	}

	colWidths := make(map[string]int)

	minWidths := make(map[string]int)
	for _, header := range m.headers {
		minWidths[header] = len(header)
	}

	for _, row := range m.data {
		for _, header := range m.headers {
			if val, exists := row[header]; exists {
				strVal := fmt.Sprintf("%v", val)
				if len(strVal) > minWidths[header] {
					minWidths[header] = len(strVal)
				}
			}
		}
	}

	totalMinWidth := 0
	for _, width := range minWidths {
		totalMinWidth += width
	}

	if totalMinWidth > 0 {
		for _, header := range m.headers {
			proportion := float64(minWidths[header]) / float64(totalMinWidth)
			calculatedWidth := int(float64(availableWidth) * proportion)

			if calculatedWidth < 8 {
				calculatedWidth = 8
			}
			if calculatedWidth > 50 {
				calculatedWidth = 50
			}

			colWidths[header] = calculatedWidth
		}
	} else {
		equalWidth := availableWidth / len(m.headers)
		for _, header := range m.headers {
			colWidths[header] = equalWidth
		}
	}

	var b strings.Builder

	var headerCells []string
	for _, header := range m.headers {
		cell := headerStyle.Width(colWidths[header]).Render(truncateString(header, colWidths[header]))
		headerCells = append(headerCells, cell)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	visibleStart := m.viewport.start
	visibleEnd := m.viewport.end
	if visibleEnd > len(m.data) {
		visibleEnd = len(m.data)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		row := m.data[i]
		var rowCells []string

		for _, header := range m.headers {
			val := ""
			if v, exists := row[header]; exists {
				val = fmt.Sprintf("%v", v)
			}

			cellStyle := normalRowStyle
			if i%2 == 1 {
				cellStyle = altRowStyle
			}
			if i == m.cursor {
				cellStyle = selectedRowStyle
			}

			cell := cellStyle.Width(colWidths[header]).Render(truncateString(val, colWidths[header]))
			rowCells = append(rowCells, cell)
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
		b.WriteString("\n")
	}

	const helpBar = "j/k: up/down • g/G: top/bottom • q: quit • r: retrieve • d: delete"

	table := tableStyle.Render(b.String())
	table = tableStyle.Width(m.width - 2).Render(b.String())
	help := helpStyle.Render(helpBar)
	status := helpStyle.Render(fmt.Sprintf("Row %d of %d", m.cursor+1, len(m.data)))

	if m.message != "" {
		help = helpStyle.Render(m.message)
	} else {
		help = helpStyle.Render(helpBar)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		table,
		lipgloss.JoinHorizontal(lipgloss.Left, help, "  ", status),
	)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func showTable(data []map[string]interface{}) error {
	if len(data) == 0 {
		fmt.Println("No files found in vault")
		return nil
	}

	p := tea.NewProgram(newTableModel(data), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run table viewer: %w", err)
	}
	return nil
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Get a list of all the files in your vault",
	Long:  "A table view of all the files that are in your encrypted vault",
	Run: func(cmd *cobra.Command, args []string) {

		data, err := utils.GetVaultContent(authenticatedPassword)

		if err != nil {
			color.Yellow("Could not get vault content")
			return
		}

		files := []map[string]interface{}{}

		for _, file := range data {
			fileMap := map[string]interface{}{
				"Id":            file.Id,
				"Original Name": file.OriginalName,
				"Date Added":    file.DateAdded.Format("2006-01-02 15:04:05"),
				"Mime Type":     file.MimeType,
				"Extension":     file.Extension,
			}
			files = append(files, fileMap)
		}

		if err := showTable(files); err != nil {
			log.Fatalf("Error displaying table: %v", err)
		}
	},
}
