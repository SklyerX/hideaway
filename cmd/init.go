package cmd

import (
	"encoding/json"
	"fmt"
	"hideaway/utils"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type model struct {
	inputs     []textinput.Model
	focusIndex int
	done       bool
}

func initialModel() model {
	inputs := make([]textinput.Model, 2)

	// Password input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Enter password"
	inputs[0].EchoMode = textinput.EchoPassword
	inputs[0].Focus()
	inputs[0].Width = 30

	// Confirm password input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Confirm password"
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].Width = 30

	return model{
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			m.inputs[m.focusIndex].Blur()
			m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
			m.inputs[m.focusIndex].Focus()
		case tea.KeyEnter:
			if (len(m.inputs[0].Value()) > 0 && len(m.inputs[1].Value()) > 0) && m.inputs[0].Value() == m.inputs[1].Value() {
				m.done = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Border(lipgloss.ASCIIBorder()).Padding(1, 4).Bold(true)

	s := titleStyle.Render("🔐 Hideaway Setup") + "\n\n"
	s += "🔐 Setup Master Password\n\n"
	s += "Password:\n"
	s += m.inputs[0].View() + "\n\n"

	s += "Confirm Password:\n"
	s += m.inputs[1].View() + "\n\n"

	return s
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Hideaway",
	Long:  "Initialize your settings and setup hideaway",
	Run: func(cmd *cobra.Command, args []string) {
		filePaths := utils.GetAppPaths()
		configFilePath := filepath.Join(filePaths["userData"], "config.json")

		if _, err := os.Stat(configFilePath); err == nil {
			fmt.Println("You have already initialized Hideaway, if you're looking to reset your app run 'hideaway reset'")
			return
		} else if !os.IsNotExist(err) {
			fmt.Printf("Error checking config file: %v\n", err)
			return
		}

		p := tea.NewProgram(initialModel())
		finalModel, err := p.Run()

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		m := finalModel.(model)
		if !m.done {
			fmt.Println("\n❌ Setup cancelled")
			return
		}

		salt, err := utils.GenerateSalt(32)
		if err != nil {
			fmt.Println("Something went wrong while generating salt rounds")
			return
		}

		passwordInBytes := []byte(m.inputs[0].Value())
		hashed, err := utils.Hash(passwordInBytes, salt)
		if err != nil {
			fmt.Println("Something went wrong while hashing the password")
			return
		}

		config := utils.Config{
			Salt:           salt,
			HashedPassword: hashed,
		}

		if err := createRootFolder(); err != nil {
			fmt.Println("Something went wrong while creating root folder")
			return
		}
		if err := createConfig(config); err != nil {
			fmt.Println("Something went wrong while creating config file")
			return
		}

		fmt.Println("Successfully initialized Hideaway, run 'hideaway' to launch the app")
		return

	},
}

func createConfig(data utils.Config) error {
	filePaths := utils.GetAppPaths()
	configFilePath := filepath.Join(filePaths["userData"], "config.json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func createRootFolder() error {
	filePaths := utils.GetAppPaths()
	rootPath := filePaths["userData"]

	if _, err := os.Stat(rootPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	err := os.Mkdir(rootPath, 0755)

	if err != nil {
		fmt.Printf("Something went wrong while creating root folder\n%v", err)
		return err
	}

	err = os.Mkdir(fmt.Sprintf("%s/dump", rootPath), 0755)

	if err != nil {
		fmt.Printf("Something went wrong while creating the files folder\n%v", err)
		return err
	}

	return nil
}
