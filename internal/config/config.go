package config

type Config struct {
	General `mapstructure:"general"`
	Rest    `mapstructure:"rest"`
	P2P     `mapstructure:"p2p"`
	Job     `mapstructure:"job"`
}

type General struct {
	MetadataPath string `mapstructure:"metadata_path"`
	Debug        bool   `mapstructure:"debug"`
}

type Rest struct {
	Port int `mapstructure:"port"`
}

type P2P struct {
	ListenAddress  []string `mapstructure:"listen_address"`
	BootstrapPeers []string `mapstructure:"bootstrap_peers"`
}

type Job struct {
	GistUpdateInterval int `mapstructure:"gist_update_interval"` // in minutes
}
