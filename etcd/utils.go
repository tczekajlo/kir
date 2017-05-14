package etcd

import (
	"log"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/viper"
)

func authCfgFromCmd() *authCfg {
	var err error
	userFlag := viper.GetString("etcd.user")

	if userFlag == "" {
		return nil
	}

	var cfg authCfg

	splitted := strings.SplitN(userFlag, ":", 2)
	if len(splitted) < 2 {
		cfg.username = userFlag
		cfg.password, err = speakeasy.Ask("Password: ")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cfg.username = splitted[0]
		cfg.password = splitted[1]
	}

	return &cfg
}
