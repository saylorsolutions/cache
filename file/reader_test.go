package file

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type _testData struct {
	Name string `json:"name"`
	Desc string `json:"description"`
}

func TestNewReaderCache(t *testing.T) {
	tmp, err := os.MkdirTemp("", "NewReaderCache-*")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmp))
	}()
	filename := filepath.Join(tmp, "test.txt")
	require.NoError(t, os.WriteFile(filename, []byte(`{"name":"Go","description":"A super cool language and ecosystem"}`), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache, err := NewReaderCache[_testData](ctx, filename, func(reader io.Reader) (_testData, error) {
		var data _testData
		err := json.NewDecoder(reader).Decode(&data)
		return data, err
	}, testingLog(t))
	assert.NoError(t, err)

	data, err := cache.Get()
	assert.NoError(t, err)
	assert.Equal(t, _testData{Name: "Go", Desc: "A super cool language and ecosystem"}, data)
}
