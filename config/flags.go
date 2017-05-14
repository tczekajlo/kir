package config

const (
	// EtcdEndpoints contains the name of flag to pass etcd endpoints
	EtcdEndpoints string = "etcd-endpoints"

	// EtcdPrefix contains the name of flag to pass etcd prefix path
	EtcdPrefix string = "etcd-prefix"

	// EtcdGetLimit is a limit of returns result in
	// the case when are get all rules from etcd
	EtcdGetLimit int64 = 20
)
