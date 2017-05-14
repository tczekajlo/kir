package etcd

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/pkg/capnslog"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/tczekajlo/kir/pb"
	"golang.org/x/net/context"
)

var plog = capnslog.NewPackageLogger("github.com/coreos/etcd", "clientv3")
var logLevel = capnslog.DEBUG

type Client struct {
	Client         *clientv3.Client
	GetResponse    *clientv3.GetResponse
	DeleteResponse *clientv3.DeleteResponse
	TxnResponse    *clientv3.TxnResponse
}

func newClientCfg(endpoints []string, dialTimeout time.Duration, scfg *secureCfg, acfg *authCfg) (*clientv3.Config, error) {
	// set tls if any one tls option set
	var cfgtls *transport.TLSInfo
	tlsinfo := transport.TLSInfo{}
	if scfg.cert != "" {
		tlsinfo.CertFile = scfg.cert
		cfgtls = &tlsinfo
	}

	if scfg.key != "" {
		tlsinfo.KeyFile = scfg.key
		cfgtls = &tlsinfo
	}

	if scfg.cacert != "" {
		tlsinfo.CAFile = scfg.cacert
		cfgtls = &tlsinfo
	}

	cfg := &clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
	if cfgtls != nil {
		clientTLS, err := cfgtls.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = clientTLS
	}
	// if key/cert is not given but user wants secure connection, we
	// should still setup an empty tls configuration for gRPC to setup
	// secure connection.
	if cfg.TLS == nil && !scfg.insecureTransport {
		cfg.TLS = &tls.Config{}
	}

	// If the user wants to skip TLS verification then we should set
	// the InsecureSkipVerify flag in tls configuration.
	if scfg.insecureSkipVerify && cfg.TLS != nil {
		cfg.TLS.InsecureSkipVerify = true
	}

	if acfg != nil {
		cfg.Username = acfg.username
		cfg.Password = acfg.password
	}

	return cfg, nil
}

func (c *Client) New() {
	capnslog.SetGlobalLogLevel(logLevel)
	clientv3.SetLogger(plog)

	endpoints := viper.GetStringSlice("etcd.endpoints")
	dialTimeout := viper.GetDuration("etcd.dial_timeout")
	sec := &secureCfg{
		cert:   viper.GetString("etcd.cert"),
		key:    viper.GetString("etcd.key"),
		cacert: viper.GetString("etcd.cacert"),

		insecureTransport:  viper.GetBool("etcd.insecure_transport"),
		insecureSkipVerify: viper.GetBool("etcd.insecure_skip_tls_verify"),
	}

	auth := authCfgFromCmd()

	cfg, err := newClientCfg(endpoints, dialTimeout, sec, auth)
	if err != nil {
		log.Panic(err)
	}

	cli, err := clientv3.New(*cfg)
	if err != nil {
		log.Panic(err)
	}

	cli.KV = namespace.NewKV(cli.KV, viper.GetString("etcd.prefix"))
	cli.Watcher = namespace.NewWatcher(cli.Watcher, viper.GetString("etcd.prefix"))
	cli.Lease = namespace.NewLease(cli.Lease, viper.GetString("etcd.prefix"))

	c.Client = cli
}

func (c *Client) Add(data *pb.Rule, override bool) error {
	// protobuf
	out, err := proto.Marshal(data)
	if err != nil {
		log.Fatalln("Failed to encode rule:", err)
	}

	key := "rule/" + data.Name
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("etcd.command_timeout"))

	// override
	var txnCompare clientv3.Cmp
	if override {
		txnCompare = clientv3util.KeyExists(key)
	} else {
		txnCompare = clientv3util.KeyMissing(key)
	}
	c.TxnResponse, err = c.Client.Txn(ctx).
		If(txnCompare).
		Then(clientv3.OpPut(key, string(out))).
		Commit()

	cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Printf("ctx is canceled by another routine: %v\n", err)
			return err
		case context.DeadlineExceeded:
			fmt.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
			return err
		case rpctypes.ErrEmptyKey:
			fmt.Printf("client-side error: %v\n", err)
			return err
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
			return err
		}
	}
	return err
}

func (c *Client) GetAll(limit int64) (*pb.RulesList, error) {
	var rule *pb.Rule
	var err error

	result := &pb.RulesList{}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("etcd.command_timeout"))
	c.GetResponse, err = c.Client.Get(ctx, "rule", clientv3.WithLimit(limit), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	cancel()
	if err != nil {
		return nil, err
	}

	for _, ev := range c.GetResponse.Kvs {
		rule = &pb.Rule{}
		if err := proto.Unmarshal(ev.Value, rule); err != nil {
			return nil, fmt.Errorf("Failed to parse rule: %s", err)
		}

		result.Rule = append(result.Rule, rule)
	}
	return result, nil
}

func (c *Client) Get(key string) (*pb.Rule, error) {
	var err error
	rule := &pb.Rule{}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("etcd.command_timeout"))
	c.GetResponse, err = c.Client.Get(ctx, key)
	cancel()
	if err != nil {
		log.Fatal(err)
	}

	if c.GetResponse.Count == 0 {
		return rule, fmt.Errorf("Cannot find rule")
	}

	for _, ev := range c.GetResponse.Kvs {
		rule = &pb.Rule{}
		if err := proto.Unmarshal(ev.Value, rule); err != nil {
			log.Fatalln("Failed to parse rule:", err)
		}

	}
	return rule, nil
}

func (c *Client) Delete(key string) error {
	var err error

	key = "rule/" + key
	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("etcd.command_timeout"))
	c.DeleteResponse, err = c.Client.Delete(ctx, key)
	cancel()
	if err != nil {
		return err
	}

	return nil
}
