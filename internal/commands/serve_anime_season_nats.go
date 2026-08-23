package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/anime-sync/internal/eventing"
)

// serveAnimeSeasonNatsCmd is the NATS counterpart of serve-anime-season-kafka.
//
// A separate command rather than a flag on the Kafka one: production and
// staging run the same image, and the command name is what decides which
// transport a deployment uses. That keeps the choice in the ArgoCD values
// rather than in an environment variable that is easy to get wrong.
var serveAnimeSeasonNatsCmd = &cobra.Command{
	Use:   "serve-anime-season-nats",
	Short: "Consume anime season change events from NATS JetStream",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running anime season eventing over NATS...")

		return eventing.EventingAnimeSeasonNats()
	},
}

func init() {
	rootCmd.AddCommand(serveAnimeSeasonNatsCmd)
}
