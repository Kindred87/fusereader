package fusereader

import (
	"context"
	"testing"
)

func Test_readWorker_run(t *testing.T) {
	readBuf := make(chan readRow)
	go consumeReadBuffer(readBuf)
	defer close(readBuf)

	workPool := make(chan int, 10)
	for i := 0; i < cap(workPool); i++ {
		workPool <- 1
	}
	defer close(workPool)

	tests := []struct {
		name    string
		r       *readWorker
		wantErr bool
	}{
		{name: "Valid", r: &readWorker{file: fuseTestFiles[1], sheet: worksheetFSItem, ctx: context.Background(), readBuffer: readBuf, workerPool: workPool}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.run(); (err != nil) != tt.wantErr {
				t.Errorf("readWorker.run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// consumeReadBuffer empties the given buffer until it is closed.
func consumeReadBuffer(c chan readRow) {
	for {
		_, ok := <-c

		if !ok {
			return
		}
	}
}
