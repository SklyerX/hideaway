package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	utils "github.com/sklyerx/hideaway/utils"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authenticatedPassword []byte

var rootCmd = &cobra.Command{
	Use:   "hideaway",
	Short: "Secure file encryption and storage",
	Long: `Hideaway encrypts and stores your files securely using a master password.
Run 'hideaway init' to set up, then 'hideaway' to enter interactive mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !isInitialized() {
			fmt.Println("Hideaway has not been initialized yet. Run 'hideaway init' first.")
			return
		}

		fmt.Print("Enter password: ")

		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("Error reading password:", err)
			return
		}

		fmt.Println()

		c, err := utils.ReadConfig()
		if err != nil {
			fmt.Println("Error while reading config")
			return
		}

		valid, err := utils.VerifyPassword(password, c.HashedPassword, c.Salt)

		if err != nil {
			fmt.Println("Error verifying password:", err)
			return
		}

		if !valid {
			fmt.Println("Invalid password.")
			return
		}

		authenticatedPassword = password

		startRepl()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(resetCmd)
}

func isInitialized() bool {
	filePaths := utils.GetAppPaths()
	configFilePath := filepath.Join(filePaths["userData"], "config.json")

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}

	return true
}

func startRepl() {
	fmt.Println("Welcome to Hideaway Repl!")
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")

	playgroundCmd := createPlaygroundCommands()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" || input == "bye" {
			fmt.Println("Bye!")
			return
		}

		// Use shell-like parsing instead of strings.Fields
		args, err := parseShellArgs(input)
		if err != nil {
			fmt.Printf("Error parsing command: %v\n", err)
			continue
		}

		playgroundCmd.SetArgs(args)
		playgroundCmd.SetOut(os.Stdout)
		playgroundCmd.SetErr(os.Stderr)

		if err := playgroundCmd.Execute(); err != nil {
			if strings.Contains(err.Error(), "unknown command") {
				fmt.Printf("Unknown command: %s\n", args[0])
				fmt.Println("Type 'help' for available commands.")
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		}

		fmt.Println()
	}
}

func parseShellArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	for i, char := range input {
		switch {
		case !inQuotes && (char == '"' || char == '\''):
			inQuotes = true
			quoteChar = char
		case inQuotes && char == quoteChar:
			inQuotes = false
		case !inQuotes && char == ' ':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			for i+1 < len(input) && input[i+1] == ' ' {
				i++
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func createPlaygroundCommands() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statsCmd)

	addCmd.Flags().BoolP("delete", "d", false, "Delete the original file after storing the encrypted version")
	addCmd.Flags().StringP("name", "n", "", "Add your own custom name (instead of the program interpreting the original file name) for better organization")

	rootCmd.PersistentFlags().ParseErrorsWhitelist.UnknownFlags = true
	addCmd.Flags().SetInterspersed(true)

	return rootCmd
}
