package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hideaway/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset your Hideaway vault",
	Long:  "A complete and full reset of Hideaway",
	Run: func(cmd *cobra.Command, args []string) {
		var confirmation string

		filePaths := utils.GetAppPaths()

		configPath := filepath.Join(filePaths["userData"], "config.json")
		databasePath := filepath.Join(filePaths["userData"], "db.enc")
		dumpPath := filepath.Join(filePaths["userData"], "dump")

		color.Red("Are you absolutely sure you want to reset your Hideaway vault?")

		color.Yellow(`
By resetting your Hideaway vault all information will be lost such as: Encrypted files, metadata, database, and master password hash
		`)

		color.Red("This action is NOT reversible\n")

		fmt.Print("If you are absolutely certain about this action type 'confirm': ")

		fmt.Scanln(&confirmation)

		if confirmation != "confirm" {
			fmt.Print("Invalid confirmation message, aborting...")

			return
		}

		os.Remove(configPath)
		os.Remove(databasePath)
		os.RemoveAll(dumpPath)

		color.Cyan("Successfully reset Hideaway, if you wish to continue using Hideaway run 'hideaway init'")
	},
}
