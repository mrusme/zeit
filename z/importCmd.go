package z

import (
	"fmt"
	"os"
	"time"

	"github.com/cnf/structhash"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

func importTymeJson(user string, file string) ([]Entry, error) {
	var entries []Entry

	tyme := Tyme{}
	tyme.Load(file)

	for _, tymeEntry := range tyme.Data {
		tymeEntrySHA1 := structhash.Sha1(tymeEntry, 1)
		tymeStart, err := time.Parse("2006-01-02T15:04:05-07:00", tymeEntry.Start)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			continue
		}

		tymeEnd, err := time.Parse("2006-01-02T15:04:05-07:00", tymeEntry.End)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			continue
		}

		entry, err := NewEntry("", "", "", tymeEntry.Project, tymeEntry.Task, user)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			continue
		}

		entry.Begin = tymeStart
		entry.Finish = tymeEnd

		entry.SHA1 = fmt.Sprintf("%x", tymeEntrySHA1)

		entries = append(entries, entry)
	}

	return entries, nil
}

var importCmd = &cobra.Command{
	Use:   "import ([flags]) [file]",
	Short: "Import tracked activities",
	Long:  "Import tracked activities from various formats.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []Entry
		var err error

		user := GetCurrentUser()

		switch format {
		case "zeit":
			// TODO:
			fmt.Printf("%s not yet implemented\n", CharError)
			os.Exit(1)
		case "tyme":
			entries, err = importTymeJson(user, args[0])
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		default:
			fmt.Printf("%s specify an import format; see `zeit import --help` for more info\n", CharError)
			os.Exit(1)
		}

		sha1List, sha1Err := database.GetImportsSHA1List(user)
		if sha1Err != nil {
			fmt.Printf("%s %+v\n", CharError, sha1Err)
			os.Exit(1)
		}

		for _, entry := range entries {
			if id, ok := sha1List[entry.SHA1]; ok {
				fmt.Printf("%s %s was previously imported as %s; not importing again\n", CharInfo, color.FgLightWhite.Render(entry.SHA1), color.FgLightWhite.Render(id))
				continue
			}

			importedId, err := database.AddEntry(user, entry, false)
			if err != nil {
				fmt.Printf("%s %s could not be imported: %+v\n", CharError, color.FgLightWhite.Render(entry.SHA1), color.FgRed.Render(err))
				continue
			}

			fmt.Printf("%s %s was imported as %s\n", CharInfo, color.FgLightWhite.Render(entry.SHA1), color.FgLightWhite.Render(importedId))
			sha1List[entry.SHA1] = importedId
		}

		err = database.UpdateImportsSHA1List(user, sha1List)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		return
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&format, "format", "zeit", "Format to import, possible values: zeit, tyme")
}
