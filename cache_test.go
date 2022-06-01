package fusereader

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheInMemory(t *testing.T) {
	tests := []struct {
		name     string
		want     Option
		expectID optionID
	}{
		{name: "Basic", want: optionCacheInMemory{}, expectID: idCacheInMemory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CacheInMemory(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CacheInMemory() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, tt.expectID, CacheInMemory().id())
		})
	}
}

func TestCacheOnDisk(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		want     Option
		expectID optionID
	}{
		{want: &optionCacheOnDisk{}, expectID: idCacheOnDisk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CacheOnDisk(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CacheOnDisk() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, tt.expectID, CacheOnDisk(context.TODO()).id())
		})
	}
}

func TestCacheOnDiskCtx(t *testing.T) {
	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	cacheOpt, ok := cacheOnDiskFrom(CacheOnDisk(cancelCtx))
	require.True(t, ok)

	/*
		anyVar := cacheOpt.Args()
		fromOption, ok := anyVar.(context.Context)

		assert.True(t, ok)
		require.NotNil(t, fromOption)
	*/

	cancelFunc()

	select {
	case <-cacheOpt.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Context cancellation", "Cancel func did not propagate to the option's internal context")
	}

}
