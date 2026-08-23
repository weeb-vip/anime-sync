package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/anime-sync/internal/eventing"
)

// serveAnimeNatsCmd is the NATS counterpart of serve-anime-kafka.
//
// A separate command rather than a flag on the Kafka one: production and
// staging run the same image, and the command name is what decides which
// transport a deployment uses. That keeps the choice in the ArgoCD values
// rather than in an environment variable that is easy to get wrong.
var serveAnimeNatsCmd = &cobra.Command{
	Use:   "serve-anime-nats",
	Short: "Consume anime change events from NATS JetStream",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running anime eventing over NATS...")

		return eventing.EventingAnimeNats()
	},
}

func init() {
	rootCmd.AddCommand(serveAnimeNatsCmd)
}
