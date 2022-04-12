package tsdb

import "flag"

type Config struct {
	Store string `yaml:"store"`
}

func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.Store, "tsdb.store", "", "Shared store for keeping tsdb files. Supported types: gcs, s3, azure, filesystem")
}
