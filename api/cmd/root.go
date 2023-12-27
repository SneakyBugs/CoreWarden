package cmd

import (
	"fmt"
	"os"
	"strings"

	"git.houseofkummer.com/lior/home-dns/api/services"
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
		Use:   "api",
		Short: "DNS management API server",
		Long: "API server for managing DNS records.\n\n" +
			"Flags can be set through config file. Supports reading\n" +
			"from JSON, TOML, YAML, and HCL files.\n\n" +
			"Flags can be set through environment variables prefixed\n" +
			"with 'DNSAPI_', all uppercase, with '-' replaced by '_'.\n" +
			"For example:\n" +
			"  postgres-host flag becomes DNSAPI_POSTGRES_HOST",
		Run: func(cmd *cobra.Command, args []string) {
			app := services.NewApp(services.Options{
				GRPCPort:         cfg.GetUint16("grpc-port"),
				HTTPPort:         cfg.GetUint16("http-port"),
				PostgresDatabase: cfg.GetString("postgres-database"),
				PostgresHost:     cfg.GetString("postgres-host"),
				PostgresPassword: cfg.GetString("postgres-password"),
				PostgresPort:     cfg.GetUint16("postgres-port"),
				PostgresUser:     cfg.GetString("postgres-user"),
				PolicyFile:       cfg.GetString("policy-file"),
			})
			app.Run()
		},
	}

	var configFile string
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config path (default dns-api.*)")

	cmd.Flags().String("policy-file", "", "Casbin policy CSV file (default policy.csv)")
	_ = cfg.BindPFlag("policy-file", cmd.Flags().Lookup("policy-file"))
	cfg.SetDefault("policy-file", "policy.csv")

	cmd.Flags().Uint16("grpc-port", 0, "gRPC resolver server listen port (default 6969)")
	_ = cfg.BindPFlag("grpc-port", cmd.Flags().Lookup("grpc-port"))
	cfg.SetDefault("grpc-port", 6969)

	cmd.Flags().Uint16("http-port", 0, "HTTP REST server listen port (default 6970)")
	_ = cfg.BindPFlag("http-port", cmd.Flags().Lookup("http-port"))
	cfg.SetDefault("http-port", 6970)

	cmd.Flags().String("postgres-host", "", "Postgres host address")
	_ = cfg.BindPFlag("postgres-host", cmd.Flags().Lookup("postgres-host"))

	cmd.Flags().Uint16("postgres-port", 0, "Postgres port (default 5432)")
	_ = cfg.BindPFlag("postgres-port", cmd.Flags().Lookup("postgres-port"))
	cfg.SetDefault("postgres-port", 5432)

	cmd.Flags().String("postgres-database", "", "Postgres database")
	_ = cfg.BindPFlag("postgres-database", cmd.Flags().Lookup("postgres-database"))

	cmd.Flags().String("postgres-user", "", "Postgres user")
	_ = cfg.BindPFlag("postgres-user", cmd.Flags().Lookup("postgres-user"))

	cmd.Flags().String("postgres-password", "", "Postgres password")
	_ = cfg.BindPFlag("postgres-password", cmd.Flags().Lookup("postgres-password"))

	cobra.OnInitialize(func() {
		if configFile != "" {
			cfg.SetConfigFile(configFile)
		} else {
			cfg.AddConfigPath(".")
			cfg.SetConfigName("dns-api")
		}

		cfg.SetEnvPrefix("dnsapi")
		cfg.AutomaticEnv()

		if err := cfg.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("no config found")
				return
			}
			fmt.Println(err)
			os.Exit(1)
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
