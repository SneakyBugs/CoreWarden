package cmd

import (
	"fmt"
	"os"
	"strings"

	"git.houseofkummer.com/lior/home-dns/external-dns/api/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var envReplacer = strings.NewReplacer("-", "_")

type Command struct {
	Cmd *cobra.Command
	Cfg *viper.Viper
}

func CreateRootCommand() Command {
	cfg := viper.NewWithOptions(
		viper.EnvKeyReplacer(envReplacer),
	)

	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "ExternalDNS webhook for DNS API",
		Long: "Webhook for managing DNS records automatically with ExternalDNS.\n\n" +
			"Flags can be set through config file. Supports reading\n" +
			"from JSON, TOML, YAML, and HCL files.\n\n" +
			"Flags can be set through environment variables prefixed\n" +
			"with 'WEBHOOK_', all uppercase, with '-' replaced by '_'.\n" +
			"For example:\n" +
			"  --api-endpoint flag becomes WEBHOOK_API_ENDPOINT",
		Run: func(cmd *cobra.Command, args []string) {
			app := services.NewApp(services.Options{
				APIEndpoint: cfg.GetString("api-endpoint"),
				ID:          cfg.GetString("api-id"),
				Secret:      cfg.GetString("api-secret"),
				Zones:       cfg.GetString("zones"),
				Port:        cfg.GetUint16("port"),
			})
			app.Run()
		},
	}

	var configFile string
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config path (default webhook.*)")

	cmd.Flags().String("api-endpoint", "", "API server URL, for example 'https://dns.example.com/v1'")
	_ = cfg.BindPFlag("api-endpoint", cmd.Flags().Lookup("api-endpoint"))

	cmd.Flags().String("api-id", "", "ID for API server authentication")
	_ = cfg.BindPFlag("api-id", cmd.Flags().Lookup("api-id"))

	cmd.Flags().String("api-secret", "", "Secret for API server authentication")
	_ = cfg.BindPFlag("api-secret", cmd.Flags().Lookup("api-secret"))

	cmd.Flags().String("zones", "", "Comma-separated list of zones managed by the webhook, for example 'example.com.,example.net.'")
	_ = cfg.BindPFlag("zones", cmd.Flags().Lookup("zones"))

	cmd.Flags().Uint16("port", 0, "HTTP server listen port (default 8888)")
	_ = cfg.BindPFlag("port", cmd.Flags().Lookup("port"))
	cfg.SetDefault("port", 8888)

	cobra.OnInitialize(func() {
		if configFile != "" {
			cfg.SetConfigFile(configFile)
		} else {
			cfg.AddConfigPath(".")
			cfg.SetConfigName("webhook")
		}

		cfg.SetEnvPrefix("webhook")
		cfg.AutomaticEnv()

		if err := cfg.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		missingOptions := []string{}
		for _, key := range cfg.AllKeys() {
			if !cfg.IsSet(key) {
				missingOptions = append(missingOptions, key)
			}
		}
		if 0 < len(missingOptions) {
			fmt.Printf("Error:\n  Missing required arguments: %v\n\n", missingOptions)
			_ = cmd.Usage()
			os.Exit(1)
		}
	})

	return Command{
		Cmd: cmd,
		Cfg: cfg,
	}
}
