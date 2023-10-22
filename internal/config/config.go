package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccuralSystemAddress string
}

func NewConfig() *Config {
	config := Config{}
	config.setConfig()
	return &config
}

func (c *Config) setConfig() {
	//get variables from flags
	flag.StringVar(&c.RunAddress, "a", "localhost:8080", "adress and port for start server")
	flag.StringVar(&c.DatabaseURI, "d", "", "for database link")
	flag.StringVar(&c.AccuralSystemAddress, "r", "", "accural_system_address")
	flag.Parse()

	//check global variables
	runAddress, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		c.RunAddress = runAddress
	}
	databaseURI, ok := os.LookupEnv("DATABASE_URI")
	if ok {
		c.DatabaseURI = databaseURI
	}
	accuralSystemAdress, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		c.AccuralSystemAddress = accuralSystemAdress
	}

}
