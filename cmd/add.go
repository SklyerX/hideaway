package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	utils "github.com/sklyerx/hideaway/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file>",
	Short: "Add a file to Hideaway",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteOriginal, _ := cmd.Flags().GetBool("delete")
		newName, _ := cmd.Flags().GetString("name")

		path := args[0]

		if strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'") {
			path = strings.TrimPrefix(path, "'")
			path = strings.TrimSuffix(path, "'")
		} else if strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"") {
			path = strings.TrimPrefix(path, "\"")
			path = strings.TrimSuffix(path, "\"")
		}

		path = strings.ReplaceAll(path, "\\ ", " ")
		path = strings.ReplaceAll(path, "\\[", "[")
		path = strings.ReplaceAll(path, "\\]", "]")
		path = strings.ReplaceAll(path, "\\(", "(")
		path = strings.ReplaceAll(path, "\\)", ")")
		path = strings.ReplaceAll(path, "\\&", "&")

		path = utils.CleanPath(path)

		dir := filepath.Dir(path)
		targetName := filepath.Base(path)

		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading directory: %v", err)
			return
		}

		var actualPath string
		for _, entry := range entries {
			cleanedEntryName := utils.CleanPath(entry.Name())
			if cleanedEntryName == targetName {
				actualPath = filepath.Join(dir, entry.Name())
				break
			}
		}

		if actualPath == "" {
			fmt.Printf("File not found: %s", path)
			return
		}

		filePaths := utils.GetAppPaths()
		dumpPath := filepath.Join(filePaths["userData"], "dump")

		err, data := utils.EncryptFile(actualPath, dumpPath, newName, string(authenticatedPassword))

		if err != nil {
			fmt.Printf("Something went wrong while encrypting file: %v", err)
			return
		}

		if deleteOriginal {
			os.Remove(actualPath)
			color.Red(fmt.Sprintf("[ DELETED ] file '%s' from disk (stored in vault)", data.OriginalName))
		}

		err = utils.AppendFile(data, authenticatedPassword)

		if err != nil {
			color.Yellow("Something went wrong while handling vault update")
			return
		}
	},
}
