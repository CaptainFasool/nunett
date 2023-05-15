package config

import (
	"reflect"

	"github.com/spf13/viper"
)

var cfg Config

func getViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("dms_config")
	v.SetConfigType("json")
	v.AddConfigPath(".")            // config file reading order starts with current working directory
	v.AddConfigPath("$HOME/.nunet") // then home directory
	v.AddConfigPath("/etc/nunet/")  // finally /etc/nunet
	return v
}

func setDefaultConfig() *viper.Viper {
	v := getViper()
	v.SetDefault("general.metadata_path", "/etc/nunet")
	v.SetDefault("general.debug", false)
	v.SetDefault("rest.port", 9999)
	v.SetDefault("p2p.listen_address", []string{
		"/ip4/0.0.0.0/tcp/9000",
		"/ip4/0.0.0.0/udp/9000/quic",
	})
	v.SetDefault("p2p.bootstrap_peers", []string{
		"/dnsaddr/bootstrap.p2p.nunet.io/p2p/QmQ2irHa8aFTLRhkbkQCRrounE4MbttNp8ki7Nmys4F9NP",
		"/dnsaddr/bootstrap.p2p.nunet.io/p2p/Qmf16N2ecJVWufa29XKLNyiBxKWqVPNZXjbL3JisPcGqTw",
		"/dnsaddr/bootstrap.p2p.nunet.io/p2p/QmTkWP72uECwCsiiYDpCFeTrVeUM9huGTPsg3m6bHxYQFZ",
	})
	v.SetDefault("job.gist_update_interval", 2)
	return v
}

func LoadConfig() {
	v := setDefaultConfig()
	v.ReadInConfig()
	err := v.Unmarshal(&cfg)
	if err != nil {
		// error unmarshalling config file - using default config
		setDefaultConfig().Unmarshal(&cfg)
	}
}

func SetConfig(key string, value interface{}) {
	v := getViper()
	v.Set(key, value)
	err := v.Unmarshal(&cfg)
	if err != nil {
		// error unmarshalling config file - using default config
		setDefaultConfig().Unmarshal(&cfg)
	}
}

func GetConfig() *Config {
	if reflect.DeepEqual(cfg, Config{}) {
		LoadConfig()
	}
	return &cfg
}
