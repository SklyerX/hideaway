package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hideaway/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file>",
	Short: "Add a file to Hideaway",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteOriginal, _ := cmd.Flags().GetBool("delete")

		path := strings.ReplaceAll(args[0], "'", "")
		path = strings.ReplaceAll(path, "\"", "")

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("File not found: %s", path)
			return
		} else if err != nil {
			fmt.Printf("Error accessing file %s: %v", path, err)
			return
		}

		filePaths := utils.GetAppPaths()
		dumpPath := filepath.Join(filePaths["userData"], "dump")

		err, data := utils.EncryptFile(path, dumpPath, string(authenticatedPassword))

		if err != nil {
			fmt.Printf("Something went wrong while encrypting file: %v", err)
			return
		}

		if deleteOriginal {
			os.Remove(path)
			color.Red(fmt.Sprintf("[ DELETED ] file '%s' from disk (stored in vault)", data.OriginalName))
		}

		err = utils.AppendFile(data, authenticatedPassword)

		if err != nil {
			color.Yellow("Something went wrong while handling vault update")
			return
		}
	},
}
