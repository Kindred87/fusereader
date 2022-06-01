package fusereader

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)

const (
	readBufferSendTimeoutMilliseconds    = 500
	workerPoolReceiveTimeoutMilliseconds = 1
)

var (
	ConcurrentFiles        int = 0  // ConcurrentFiles stores the number of files that fusereader will read at a time.  The default is the number of logical CPU cores.
	ReadFileTimeoutSeconds int = 30 // ReadFileTimeoutSeconds is the number of seconds that a file will be allowed to be read for before timing out.  Default is 30.
)

type readRow struct {
	file     string
	row      int
	item     string
	contents []string
}

type readWorker struct {
	file       string
	sheet      string
	ctx        context.Context
	readBuffer chan readRow
	workerPool chan int
}

func (r *readWorker) run() (err error) {
	defer func() {
		select {
		case <-r.workerPool:
		case <-time.After(workerPoolReceiveTimeoutMilliseconds * time.Millisecond):
		}
	}()

	fi, err := excelize.OpenFile(r.file)
	if err != nil {
		return fmt.Errorf("error while opening %s: %w", filepath.Base(r.file), err)
	}
	defer fi.Close()

	rows, err := fi.Rows(r.sheet)
	if err != nil {
		return fmt.Errorf("error while initiating row iterator for %s: %w", filepath.Base(r.file), err)
	}
	defer rows.Close()

	currRow := 0
	var passedHeaders bool
	var itemIdColumn int
	var currItem string

	for rows.Next() {
		select {
		case <-r.ctx.Done():
			return nil
		default:
		}

		currRow++

		cells, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("error while reading row %d in %s: %w", currRow, filepath.Base(r.file), err)
		}

		if !passedHeaders {
			if isHeaderRow(cells) {
				passedHeaders = true

				if i, ok := headerIndex(headerItemID, cells); ok {
					itemIdColumn = i
				}
			}
		} else if itemIdColumn != 0 && len(cells) >= itemIdColumn {
			if cells[itemIdColumn] != "" {
				currItem = cells[itemIdColumn]
			}
		}

		select {
		case r.readBuffer <- readRow{file: r.file, row: currRow, item: currItem, contents: cells}:
		case <-time.After(readBufferSendTimeoutMilliseconds * time.Millisecond):
			return fmt.Errorf("reader for %s timed out while waiting to send contents of row %d to the read buffer", filepath.Base(r.file), currRow)
		}

	}

	return nil
}

func readFrom(files []string, sheet string, ctx context.Context, readBuffer chan readRow) error {
	defer close(readBuffer)

	workerPool := make(chan int, readWorkerPoolSize())
	workers := make([]readWorker, len(files))
	var eg errgroup.Group

	for i, file := range files {
		workers[i] = readWorker{
			file:       file,
			sheet:      sheet,
			ctx:        ctx,
			readBuffer: readBuffer,
			workerPool: workerPool,
		}
	}

	for _, worker := range workers {
		workerPool <- 1
		w := worker
		eg.Go(func() error { return w.run() })
	}

	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("error while running file readers: %w", err)
	}

	return nil
}

func readWorkerPoolSize() int {
	if ConcurrentFiles == 0 {
		return runtime.NumCPU()
	} else {
		return ConcurrentFiles
	}
}
