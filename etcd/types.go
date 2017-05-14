package etcd

type secureCfg struct {
	cert   string
	key    string
	cacert string

	insecureTransport  bool
	insecureSkipVerify bool
}

type authCfg struct {
	username string
	password string
}
