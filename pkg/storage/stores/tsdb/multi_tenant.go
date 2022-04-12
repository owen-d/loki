package tsdb

import (
	"sync"

	chunk_util "github.com/grafana/loki/pkg/storage/chunk/client/util"
)

/*
tsdb/
    tenants/
            <user>/
                   # scratch is used as an initial flush point when building TSDB indices.
                   # After being successfully built, they're moved into
                   # the pending dir in the correct period.
                   scratch/
                           index-<rng>.staging
                   # Added to every checkpoint cycle and eventually compacted & flushed to storage.
                   pending/
                           <period1>/
                                   <from>-<through>-<checksum>.tsdb
                                   <from>-<through>-<checksum>.tsdb
                           <period2>/
                                   <from>-<through>-<checksum>.tsdb
                                   <from>-<through>-<checksum>.tsdb
                   # Has been compacted/flushed
                   flushed/
                           <period1>/
                                   <from>-<through>-<checksum>.tsdb
                                   <from>-<through>-<checksum>.tsdb
                           <period2>/
                                   <from>-<through>-<checksum>.tsdb
                                   <from>-<through>-<checksum>.tsdb
*/

type MultiTenant struct {
	dir string
	sync.RWMutex
	tenants map[string]*TenantTSDB
}

// Init creates the required directories
func (mt *MultiTenant) Init() error {
	return chunk_util.EnsureDirectory(mt.dir)
}

type TenantTSDB struct {
	dir string
	sync.RWMutex

	// pending
}
