package cmd

import (
	"fmt"
	"hideaway/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get stats about your current vault",
	Long:  "Get a short breakdown about your vault contents",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := utils.GetVaultContent(authenticatedPassword)

		if err != nil {
			color.Yellow("Something went wrong while getting vault content")
			return
		}

		if len(files) == 0 {
			fmt.Print("No files in vault")
			return
		}

		fmt.Printf("Total items in vault %d\n", len(files))

		for i, file := range files {
			if i == 0 {
				fmt.Println("---------------------------------------")
			}

			fmt.Printf(`
Size:			%d
MimeType: 		%s
Extension:		%s
Name: 			%s
Dated Added: 	%s
			`,
				file.Size, file.MimeType, file.Extension, file.OriginalName, file.DateAdded)

			fmt.Println()

			if i == (len(files) - 1) {
				fmt.Print("---------------------------------------")
			}
		}
	},
}
