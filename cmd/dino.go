package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tuna-os/bluefin-cli/internal/config"
	"github.com/tuna-os/bluefin-cli/internal/tui/app"
)

// The easter egg deserves a front door too.
var dinoCmd = &cobra.Command{
	Use:    "dino",
	Short:  "🦕",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(gameScreen())
	},
}

func init() {
	rootCmd.AddCommand(dinoCmd)
	app.Register(app.Action{
		ID: "dino", Icon: "🦕", Label: "Dino Run", Section: "Fun",
		Do: func() tea.Cmd { return app.Push(gameScreen()) },
	})
}

// gameScreen builds the dino runner with a config-persisted high score.
func gameScreen() app.Screen {
	return app.NewGame(viper.GetInt("game.high_score"), func(n int) {
		viper.Set("game.high_score", n)
		_ = config.Save()
	})
}
