package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Mryashbhardwaj/marketAnalysis/internal/api/handlers"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/api/routes"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/config"
	"github.com/Mryashbhardwaj/marketAnalysis/internal/domain/service"
	"github.com/spf13/cobra"
)

type serveCommand struct {
	port           int
	configFilePath string

	logger *slog.Logger
	config *config.Config
}

// NewServeCommand initializes command to serve the API
func NewServeCommand() *cobra.Command {
	s := &serveCommand{}

	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start the API server",
		Example: "marketWatch serve -p portnumber -c /path/to/config.yaml",
		RunE:    s.RunE,
		PreRunE: s.PreRunE,
	}

	// Config filepath flag
	cmd.Flags().StringVarP(&s.configFilePath, "config", "c", "", "File path for client configuration")

	cmd.Flags().IntVarP(&s.port, "port", "p", 8080, "service endpoint url")

	return cmd
}

func (s *serveCommand) PreRunE(cmd *cobra.Command, _ []string) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	s.logger = logger

	if s.configFilePath == "" {
		return errors.New("config file path is required")
	}

	cfg, err := config.LoadConfig(s.configFilePath)
	if err != nil {
		logger.Error("failed to open config file", slog.String("error", err.Error()))
		return err
	}

	s.config = cfg

	return err
}

func (s *serveCommand) RunE(_ *cobra.Command, _ []string) error {

	//  get tradebook

	tradebook, err := service.GetTradebookService(
		s.config.Equity.TradeFilesDirectory,
		s.config.MutualFunds.TradeFilesDirectory,
		s.logger)

	if err != nil {
		s.logger.Error("failed to get tradebook service", slog.String("error", err.Error()))
		return err
	}

	mfTrendCache := service.GetMFTrendCache(s.logger, tradebook.MutualFundsTradebookCache.ISINToFundName)
	equityTrendCache := service.GetEquityTrendCache(s.logger, tradebook.EquityTradebookCache.AllShares)

	handlers := handlers.GetHandler(tradebook, equityTrendCache, mfTrendCache)

	router := routes.SetupRouter(handlers)
	//  todo: take handlers as new handler and inject logger in handlers.SetupRouter

	// initialise server
	host := ""
	addr := fmt.Sprintf("%s:%d", host, s.port)
	s.logger.Info("Starting server ", slog.String("addr", addr))
	err = http.ListenAndServe(addr, router)
	if err != nil {
		s.logger.Error("Failed to start HTTP server", slog.String("error", err.Error()))
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
