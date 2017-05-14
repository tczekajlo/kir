package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	apiv1 "github.com/tczekajlo/kir/api/v1"
	"github.com/tczekajlo/kir/utils"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs server",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		fmt.Printf("%s\n\n", utils.Banner)

		// HTTP server
		gin.SetMode(gin.ReleaseMode)

		route := gin.Default()

		apiv1.Group(route)

		if viper.GetBool("server.tls.enabled") {
			server := endless.NewServer(viper.GetString("server.listen"), route)
			server.TLSConfig = configureTLS()

			err = server.ListenAndServeTLS(
				viper.GetString("server.tls.cert_file"),
				viper.GetString("server.tls.key_file"))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = endless.ListenAndServe(viper.GetString("server.listen"), route)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func configureTLS() *tls.Config {
	config := &tls.Config{}

	if viper.GetBool("server.tls.require_and_verify_client_cert") {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if viper.GetString("server.tls.cacert_file") != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(viper.GetString("server.tls.cacert_file"))
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config.ClientCAs = caCertPool
	}

	return config
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("listen", "l", ":8080", "address with port on which server will be listen (address:port)")
	serverCmd.Flags().Bool("tls-enabled", false, "enables SSL")
	serverCmd.Flags().String("tls-cert-file", "cert.pem", "a path to the certificate file")
	serverCmd.Flags().String("tls-key-file", "key.pem", "a path to the key file")
	serverCmd.Flags().String("tls-cacert-file", "", "a path to the root CA file")
	serverCmd.Flags().Bool("tls-require-and-verify-client-cert", false, "turns on client authentication for this listener")

	// viper
	viper.BindPFlag("server.listen", serverCmd.Flags().Lookup("listen"))
	viper.BindPFlag("server.tls.require_and_verify_client_cert", serverCmd.Flags().Lookup("tls-require-and-verify-client-cert"))
	viper.BindPFlag("server.tls.enabled", serverCmd.Flags().Lookup("tls-enabled"))
	viper.BindPFlag("server.tls.cert_file", serverCmd.Flags().Lookup("tls-cert-file"))
	viper.BindPFlag("server.tls.key_file", serverCmd.Flags().Lookup("tls-key-file"))
	viper.BindPFlag("server.tls.cacert_file", serverCmd.Flags().Lookup("tls-cacert-file"))
	viper.BindPFlag("server.tls.require_and_verify_client_cert", serverCmd.Flags().Lookup("tls-require-and-verify-client-cert"))
}
