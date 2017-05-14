package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tczekajlo/kir/config"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "kir",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kir.yaml)")
	RootCmd.PersistentFlags().StringSlice(config.EtcdEndpoints, []string{"http://localhost:2379"}, "list of URLs to etcd endpoint")
	RootCmd.PersistentFlags().String("etcd-prefix", "/kir/", "etcd prefix")
	RootCmd.PersistentFlags().Duration("etcd-dial-timeout", 2*time.Second, "dial timeout for client connections")
	RootCmd.PersistentFlags().String("etcd-cacert", "", "verify certificates of TLS-enabled secure servers using this CA bundle")
	RootCmd.PersistentFlags().String("etcd-cert", "", "identify secure client using this TLS certificate file")
	RootCmd.PersistentFlags().String("etcd-key", "", "identify secure client using this TLS key file")
	RootCmd.PersistentFlags().String("etcd-user", "", "username[:password] for authentication (prompt if password is not supplied)")
	RootCmd.PersistentFlags().Duration("etcd-command-timeout", 5*time.Second, "timeout for short running command (excluding dial timeout)")
	RootCmd.PersistentFlags().Bool("etcd-insecure-skip-tls-verify", false, "skip server certificate verification")
	RootCmd.PersistentFlags().Bool("etcd-insecure-transport", true, "disable transport security for client connections")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Fatalln(err)
		}

		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".kir") // name of config file (without extension)
	}

	viper.AddConfigPath(os.Getenv("HOME")) // adding home directory as first search path
	viper.AutomaticEnv()                   // read in environment variables that match

	// default values
	viper.BindPFlag("etcd.endpoints", RootCmd.Flags().Lookup(config.EtcdEndpoints))
	viper.BindPFlag("etcd.prefix", RootCmd.Flags().Lookup(config.EtcdPrefix))
	viper.BindPFlag("etcd.dial_timeout", RootCmd.Flags().Lookup("etcd-dial-timeout"))
	viper.BindPFlag("etcd.cacert", RootCmd.Flags().Lookup("etcd-cacert"))
	viper.BindPFlag("etcd.cert", RootCmd.Flags().Lookup("etcd-cert"))
	viper.BindPFlag("etcd.key", RootCmd.Flags().Lookup("etcd-key"))
	viper.BindPFlag("etcd.user", RootCmd.Flags().Lookup("etcd-user"))
	viper.BindPFlag("etcd.command_timeout", RootCmd.Flags().Lookup("etcd-command-timeout"))
	viper.BindPFlag("etcd.insecure_skip_tls_verify", RootCmd.Flags().Lookup("etcd-insecure-skip-tls-verify"))
	viper.BindPFlag("etcd.insecure_transport", RootCmd.Flags().Lookup("etcd-insecure-transport"))

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

}
