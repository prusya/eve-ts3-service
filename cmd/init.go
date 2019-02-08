package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/prusya/eve-ts3-service/pkg/system"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "creates config.json and logs/",
	Long: `usage: eve-ts3-service init
It will create a config template and logs directory.
You must create them in order or run the service.
Make sure to fill the config with proper values.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := system.Config{
			WebServerAddress: "127.0.0.1:8083",

			TS3Address:          "127.0.0.1:10011",
			TS3User:             "serveradmin",
			TS3Password:         "",
			TS3ServerID:         1,
			TS3Whitelisted:      "true",
			TS3ReferenceGroupID: "7",
			TS3RegisterTimer:    300,

			UsersValidationEndpoint: "http://127.0.0.1:8081/api/validation/ts3",

			PgConnString: "postgres://username:password@hostaddress/dbname?sslmode=verify-full",
		}

		cj, _ := json.MarshalIndent(c, "", "  ")
		err := ioutil.WriteFile("config.json", cj, 0644)
		if err != nil {
			panic(err)
		}
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Println("Config file created at", cwd)

		// Create logs dir.
		os.Mkdir("logs", 0644)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
