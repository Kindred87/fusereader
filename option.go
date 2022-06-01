package fusereader

import (
	"context"
)

// Option represents an optional argument that enables certain features.
type Option interface {
	id() optionID
}

// optionEnabled returns true if the given option exists within the given option slice.
func optionEnabled(needle Option, haystack []Option) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, opt := range haystack {
		if opt.id() == needle.id() {
			return true
		}
	}

	return false
}

func optionIndex(needle Option, haystack []Option) (index int, ok bool) {
	for i, opt := range haystack {
		if needle.id() == opt.id() {
			return i, true
		}
	}

	return 0, false
}

// CacheInMemory enables memory caching, if supported.
func CacheInMemory() Option {
	return optionCacheInMemory{}
}

// cacheInMemoryFrom returns a cache in memory option from the given options.
//
// If the given options do not contain a cache in memory option, then the returned
// boolean will be false.
func cacheInMemoryFrom(opts ...Option) (optionCacheInMemory, bool) {
	var out optionCacheInMemory

	i, ok := optionIndex(out, opts)
	if ok {
		out = *opts[i].(*optionCacheInMemory)
	}

	return out, ok
}

type optionCacheInMemory struct {
}

func (o optionCacheInMemory) id() optionID {
	return idCacheInMemory
}

// CacheOnDisk enables disk caching, if supported.
//
// The cache will be removed upon calling the given context's cancel func.
func CacheOnDisk(ctx context.Context) Option {
	return &optionCacheOnDisk{ctx: ctx}
}

// cacheOnDiskFrom returns a cache on disk option from the given options.
//
// If the given options do not contain a cache on disk option, then the returned
// boolean will be false.
func cacheOnDiskFrom(opts ...Option) (optionCacheOnDisk, bool) {
	var out optionCacheOnDisk

	i, ok := optionIndex(out, opts)
	if ok {
		out = *opts[i].(*optionCacheOnDisk)
	}

	return out, ok
}

type optionCacheOnDisk struct {
	ctx context.Context
}

func (o optionCacheOnDisk) id() optionID {
	return idCacheOnDisk
}
