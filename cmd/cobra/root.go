package cobra

import (
	"github.com/spf13/cobra"
	"log"
)

// Path: cmd/cobra/root.go
var rootCmd = &cobra.Command{
	Use:   "cronos-tools",
	Short: "Some useful tools on cronos",
	Long:  `Some useful tools on cronos currently including: inscription tools`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("run cronos-tools")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panicln(err)
	}
}
