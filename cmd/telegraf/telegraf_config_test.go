package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
)

func TestLoadConfigurationTestModeSkipsDiskOutputBuffer(t *testing.T) {
	savedVersion := internal.Version
	internal.Version = "0.0.0"
	defer func() {
		internal.Version = savedVersion
	}()

	root := t.TempDir()
	bufferDirectory := filepath.Join(root, "buffer")
	require.NoError(t, os.WriteFile(bufferDirectory, []byte("not a directory"), 0600))

	configFile := filepath.Join(root, "telegraf.conf")
	content := `
[agent]
  buffer_strategy = "disk"
  buffer_directory = ` + strconv.Quote(bufferDirectory) + `

[[inputs.cpu]]

[[outputs.influxdb]]
  urls = ["http://localhost:8086"]
  database = "telegraf"
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0600))

	agent := &Telegraf{
		GlobalFlags: GlobalFlags{
			config: []string{configFile},
		},
	}
	_, err := agent.loadConfiguration()
	require.ErrorContains(t, err, "creating buffer failed")

	agent = &Telegraf{
		GlobalFlags: GlobalFlags{
			config: []string{configFile},
			test:   true,
		},
	}
	cfg, err := agent.loadConfiguration()
	require.NoError(t, err)
	require.Len(t, cfg.Outputs, 1)
	require.True(t, cfg.Outputs[0].Config.SkipBuffer)
	require.Equal(t, "disk_write_through", cfg.Outputs[0].Config.BufferStrategy)
}
