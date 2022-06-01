package fusereader

import (
	"context"
	"fmt"
	"time"
)

const (
	parseBufferReceiveTimeoutMilliseconds = 5000
)

var ()

func parseBufferFor(needle, header string, buffer chan readRow, cancelRead context.CancelFunc) (readRow, error) {
	for {
		select {
		case v, match := <-buffer:
			if !match {
				return readRow{}, fmt.Errorf("could not locate %s under %s", needle, header)
			}

			row, match, err := parseFor(needle, header, v)
			if err != nil {
				return readRow{}, fmt.Errorf("error while parsing row %d from %s: %w", v.row, v.file, err)
			}

			if match {
				cancelRead()
				return row, nil
			}

		case <-time.After(parseBufferReceiveTimeoutMilliseconds * time.Millisecond):
			return readRow{}, fmt.Errorf("parser timed out while waiting to receive from buffer")
		}

	}
}

func parseFor(needle, header string, haystack readRow) (readRow, bool, error) {
	var out readRow

	if needle == "" {
		return out, false, fmt.Errorf("needle is unspecified")
	} else if header == "" {
		return out, false, fmt.Errorf("header is unspecified")
	}

	if len(haystack.contents) == 0 {
		return out, false, nil
	}

	if isHeaderRow(haystack.contents) {
		cacheHeaders(haystack.file, haystack.contents)
	}

	v, ok := valueForHeader(header, haystack.file, haystack.contents)
	if !ok {
		return out, false, nil
	}

	var match bool
	if v == needle {
		out = haystack
		match = true
	} else {
		match = false
	}

	return out, match, nil
}
