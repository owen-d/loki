package abstractshipper

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

type Index interface{}

type Options struct {
	// Time after which a Sync will remove a table|user|index
	StalePeriod time.Duration
	// How often the Sync loop should run
	SyncPeriod time.Duration
}

type TableManager struct {
	mtx sync.RWMutex
	// list of tables
	tables map[string]*Table

	opts Options

	loopCancel func()
}

func (tm *TableManager) Start() error {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if tm.loopCancel != nil {
		return errors.New("TableManager already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	tm.loopCancel = cancel

	go tm.loop(ctx)
	return nil
}

func (tm *TableManager) Stop() {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if tm.loopCancel != nil {
		tm.loopCancel()
		tm.loopCancel = nil
	}
}

func (tm *TableManager) loop(ctx context.Context) {
	ticker := time.NewTicker(tm.opts.SyncPeriod)
	tm.sync(ctx) // sync once at beginning
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.sync(ctx)
		}
	}
}

func (tm *TableManager) sync(ctx context.Context) {}

func (tm *TableManager) ForEach(ctx context.Context, userID string, tableNames []string, fn func(context.Context, Index) error) error {
	g, ctx := errgroup.WithContext(ctx)
	tm.mtx.RLock()
	defer tm.mtx.RUnlock()

	for _, name := range tableNames {
		table, ok := tm.tables[name]
		if !ok {
			continue
		}

		g.Go(func() error {
			return table.ForEach(ctx, userID, fn)
		})
	}

	return g.Wait()

}

type metadata struct {
	// unixnano representation
	lastUpdated atomic.Int64
}

func (m *metadata) touch() {
	m.lastUpdated.Store(time.Now().UnixNano())
}

type Table struct {
	mtx      sync.RWMutex
	metadata metadata
	// list of users -> files
	indexSets map[string]*IndexSet
}

func (t *Table) ForEach(ctx context.Context, userID string, fn func(context.Context, Index) error) error {
	g, ctx := errgroup.WithContext(ctx)
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	// multitenant indices are stored with an empty user ID
	// before compaction. Therefore we query both
	for _, id := range []string{userID, ""} {
		x, ok := t.indexSets[id]
		if !ok {
			continue
		}

		g.Go(func() error {
			return fn(ctx, x)
		})
	}

	err := g.Wait()
	t.metadata.touch()
	return err
}

type IndexSet struct {
	// list of files
	mtx      sync.RWMutex
	metadata metadata
	indices  map[string]Index
}

func (i *IndexSet) ForEach(ctx context.Context, fn func(context.Context, Index) error) error {
	g, ctx := errgroup.WithContext(ctx)
	i.mtx.RLock()
	defer i.mtx.RUnlock()

	for k := range i.indices {
		// need to capture variable during iteration before sending
		// to another goroutine
		x := i.indices[k]

		g.Go(func() error {
			return fn(ctx, x)
		})
	}

	err := g.Wait()
	i.metadata.touch()
	return err
}
