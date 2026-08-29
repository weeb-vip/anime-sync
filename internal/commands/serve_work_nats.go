package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/weeb-vip/anime-sync/internal/eventing"
)

// serveWorkNatsCmd consumes the scraper's `work` table into the read store.
//
// No Kafka counterpart: works never existed while this service ran on Kafka,
// so there is nothing to keep compatible.
var serveWorkNatsCmd = &cobra.Command{
	Use:   "serve-work-nats",
	Short: "Consume work (manga, light novel) change events from NATS JetStream",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Running work eventing over NATS...")

		return eventing.EventingWorkNats()
	},
}

func init() {
	rootCmd.AddCommand(serveWorkNatsCmd)
}
