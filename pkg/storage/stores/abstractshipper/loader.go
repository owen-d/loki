package abstractshipper

import "io"

type Loader interface {
	Load(userID, table, file string) (Index, error)
}

type Uploader interface {
	Store(userID, table, file string, f io.ReadSeeker)
}

/*
ForEach impl:

1) RLock
	a) Get list of (table, user) tuples that need to be sync'd
	b) RUnlock
	c) If there are none needing refreshing, iterate & return.
	d) Else, send to a downloader which can collect, download, and dedupe them
2) Wait for downloader to finish
	a) Lock
	b) Update all the local (table, user) tuples with the refreshed ones from storage
	c) Unlock
3) Goto step 1
*/
