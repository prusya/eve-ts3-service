package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/prusya/eve-ts3-service/pkg/http/gorillahttp"
	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3/darfkts3service"
	"github.com/prusya/eve-ts3-service/pkg/ts3/pgts3store"
)

// runpgCmd represents the runpg command
var runpgCmd = &cobra.Command{
	Use:   "run",
	Short: "starts the service",
	Long: `usage: eve-ts3-service run
It  will start the main functionality of the service.
Make sure to run "eve-ts3-service init" and fill the config file before "run".`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup logger.
		date := time.Now().Format("2006-01-02_15-04-05")
		f, err := os.OpenFile("logs/log_"+date+".txt", os.O_WRONLY|os.O_CREATE, 0644)
		system.HandleError(err)
		defer f.Close()
		log.SetOutput(f)
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		// Prepare gracefull shutdown.
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGKILL)

		// Create shared System for services.
		sys := system.New(sigChan)

		// Connect to db.
		db, err := sqlx.Connect("postgres", sys.Config.PgConnString)
		system.HandleError(err)
		defer db.Close()

		// Create http service.
		httpService := gorillahttp.New(sys)
		defer httpService.Stop()

		// Create ts3 service.
		ts3Store := pgts3store.New(db)
		ts3Store.Init()
		ts3Service := darfkts3service.New(sys, ts3Store)
		defer ts3Service.Stop()

		ts3Service.Start()
		httpService.Start()

		// Handle graceful shutdown.
		// Actual shutdown is performed in deferred Stop calls.
		<-sigChan
	},
}

func init() {
	rootCmd.AddCommand(runpgCmd)
}
