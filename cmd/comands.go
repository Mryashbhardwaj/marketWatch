package cmd

import (
	"github.com/MakeNowJust/heredoc"

	"github.com/Mryashbhardwaj/marketAnalysis/cmd/server"
	cli "github.com/spf13/cobra"
)

func New() *cli.Command {
	cmd := &cli.Command{
		Use: "marketWatch <command> <subcommand> [flags]",
		Long: heredoc.Doc(`
			is a personal trade analysis and portfolio visualization tool 
			that helps you understand your Equity and Mutual Fund positions 
			through easy-to-navigate dashboards powered by Grafana.`),
		SilenceUsage: true,
		Example: heredoc.Doc(`
				$ marketWatch serve
				$ marketWatch fetch-trends
			`),
		Annotations: map[string]string{
			"group:core": "true",
			"help:learn": heredoc.Doc(`
				Use 'marketWatch <command> <subcommand> --help' for more information about a command.
				Read the manual at https://goto.github.io/marketWatch/
			`),
		},
	}

	// Client related commands
	cmd.AddCommand(
		server.NewServeCommand(),
	)

	return cmd
}
