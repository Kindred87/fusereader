package fusereader

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)

// fileCache stores opened files.  Use cacheFiles to populate fileCache and closeFiles to empty it.
var fileCache map[string]*excelize.File

// getFile returns a pointer to the given file, if it is cached.
func getFile(path string) (*excelize.File, error) {
	if fileCache == nil {
		return nil, fmt.Errorf("the file cache is nil")
	}

	fo, exists := fileCache[path]
	if !exists {
		return nil, fmt.Errorf("%s has not been cached", filepath.Base(path))
	}

	return fo, nil
}

// cacheFiles caches the file pointers for the given files.
func cacheFiles(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths were given")
	}

	err := cacheFile(paths[0])
	if err != nil {
		return fmt.Errorf("error while caching %s: %w", filepath.Base(paths[0]), err)
	}

	var eg errgroup.Group

	for _, path := range paths[1:] {
		p := path
		eg.Go(func() error { return cacheFile(p) })
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error while caching files: %w", err)
	}

	return nil
}

// cacheFiles caches the file pointer for the given file.
func cacheFile(path string) error {
	fi, err := excelize.OpenFile(path)
	if err != nil {
		return fmt.Errorf("error while opening %s: %w", filepath.Base(path), err)
	}

	if fileCache == nil {
		fileCache = make(map[string]*excelize.File)
	}

	fileCache[path] = fi

	return nil
}

// closeFiles closes the cached files and empties the file cache.
func closeFiles() error {
	var eg errgroup.Group

	for _, file := range fileCache {
		f := file
		eg.Go(func() error { return f.Close() })
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error while closing files: %w", err)
	}

	fileCache = nil

	return nil
}
