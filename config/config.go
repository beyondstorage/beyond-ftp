package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pengsrc/go-shared/check"
)

// A Config stores a configuration of BeyondFTP.
type Config struct {
	Service    string            `toml:"service"`
	ListenHost string            `toml:"host"`
	ListenPort int               `toml:"port"`
	PublicHost string            `toml:"public-host"`
	StartPort  int               `toml:"start-port"`
	EndPort    int               `toml:"end-port"`
	Users      map[string]string `toml:"users"`
}

// ServerSettings define all the server settings.
type ServerSettings struct {
	Service       string
	ListenHost    string     // Host to receive connections on
	ListenPort    int        // Port to listen on
	PublicHost    string     // Public IP to expose (only an IP address is accepted at this stage)
	DataPortRange *PortRange // Port Range for data connections. Random one will be used if not specified
	Users         map[string]string
}

// PortRange is a range of ports.
type PortRange struct {
	Start int // Range start
	End   int // Range end
}

// LoadConfigFromFilepath loads configuration from a specified local path.
// It returns error if file not found or decode failed.
func LoadConfigFromFilepath(p string) *Config {
	conf := &Config{}
	if p != "" {
		_, err := toml.DecodeFile(p, conf)
		check.ErrorForExit("File not found: "+p, err)
	}
	err := setDefaultValue(conf)
	check.ErrorForExit("Config check error: ", err)
	return conf
}

// setDefaultValue checks the configuration.
func setDefaultValue(c *Config) error {
	if c.ListenHost == "" {
		c.ListenHost = "0.0.0.0"
	}
	if c.ListenPort == 0 {
		// For the default value (0), We take the default port (21).
		c.ListenPort = 21
	} else if c.ListenPort == -1 {
		// For the automatic value, We let the system decide (0).
		c.ListenPort = 0
	}
	if c.PublicHost == "" {
		c.PublicHost = "127.0.0.1"
	}
	if c.StartPort == 0 {
		c.StartPort = 1024
	}
	if c.EndPort == 0 {
		c.EndPort = 65535
	}
	if c.Users == nil {
		c.Users = make(map[string]string)
		c.Users["anonymous"] = ""
	}

	return nil
}

func GetServerSetting(c *Config) *ServerSettings {
	return &ServerSettings{
		Service:    c.Service,
		ListenHost: c.ListenHost,
		ListenPort: c.ListenPort,
		PublicHost: c.PublicHost,
		DataPortRange: &PortRange{
			Start: c.StartPort,
			End:   c.EndPort,
		},
		Users: c.Users,
	}
}
