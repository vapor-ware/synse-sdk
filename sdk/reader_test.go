package sdk

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
)

// TestIsValidConfig tests validating that a file is a potential config file.
func TestIsValidConfig(t *testing.T) {
	var testTable = []struct {
		desc    string
		isValid bool
		name    string
		file    *test.FileInfo
	}{
		{
			desc:    "file is not valid -- is a directory",
			isValid: false,
			name:    "",
			file:    test.NewFileInfo("test", os.ModeDir),
		},
		{
			desc:    "file is not valid -- is a json file",
			isValid: false,
			name:    "",
			file:    test.NewFileInfo("test,json", os.ModePerm),
		},
		{
			desc:    "file is valid -- has .yml extension",
			isValid: true,
			name:    "",
			file:    test.NewFileInfo("test.yml", os.ModePerm),
		},
		{
			desc:    "file is valid -- has .yaml extension",
			isValid: true,
			name:    "",
			file:    test.NewFileInfo("test.yaml", os.ModePerm),
		},
		{
			desc:    "file is valid -- has .yml extension and name matches",
			isValid: true,
			name:    "test",
			file:    test.NewFileInfo("test.yml", os.ModePerm),
		},
		{
			desc:    "file is not valid -- has .yml extension but name does not match",
			isValid: false,
			name:    "foo",
			file:    test.NewFileInfo("test.yml", os.ModePerm),
		},
	}

	for _, testCase := range testTable {
		result := isValidConfig(testCase.file, testCase.name)
		assert.Equal(t, testCase.isValid, result, testCase.desc)
	}
}

// Test_searchDir_NoValidFiles tests getting the paths of the files that are
// valid config files from a directory. In this case, there are no valid configs.
func Test_searchDir_NoValidFiles(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Test
	configs, err := searchDir(test.TempDir, "")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(configs))
}

// Test_searchDir_NoDir tests getting the paths of the files that are
// valid config files from a directory when the directory doesn't exist.
func Test_searchDir_NoDir(t *testing.T) {
	configs, err := searchDir("a/b/c/d/e", "")
	assert.Error(t, err)
	assert.Nil(t, configs)
}

// Test_searchDir_OneFile tests getting the paths of the files that are
// valid config files from a directory. In this case, there is only one valid config
// file, and some invalid.
func Test_searchDir_OneFile(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	foo := test.WriteTempFile(t, "foo.yml", "", 0666)
	_ = test.WriteTempFile(t, "bar.json", "", 0666)

	// Test
	configs, err := searchDir(test.TempDir, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	assert.Equal(t, foo, configs[0])
}

// Test_searchDir_OneFileInvalid tests getting the paths of the files that are
// valid config files from a directory. In this case, there are no valid config
// files, both because of extension and name.
func Test_searchDir_OneFileInvalid(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	_ = test.WriteTempFile(t, "foo.yml", "", 0666)
	_ = test.WriteTempFile(t, "bar.json", "", 0666)

	// Test
	configs, err := searchDir(test.TempDir, "config")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(configs))
}

// Test_searchDir_MultipleFiles tests getting the paths of the files that are
// valid config files from a directory. In this case, is multiple valid config files.
func Test_searchDir_MultipleFiles(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	foo := test.WriteTempFile(t, "foo.yml", "", 0666)
	bar := test.WriteTempFile(t, "bar.yaml", "", 0666)
	baz := test.WriteTempFile(t, "baz.yaml", "", 0666)

	// Test
	configs, err := searchDir(test.TempDir, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(configs))
	assert.Equal(t, bar, configs[0])
	assert.Equal(t, baz, configs[1])
	assert.Equal(t, foo, configs[2])
}

// Test_unmarshalConfigFile tests unmarshalling data from a file into a struct successfully.
func Test_unmarshalConfigFile(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
version: 3
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &DeviceConfig{}
	err := unmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, 3, config.Version)
}

// Test_unmarshalConfigFile2 tests unmarshalling data from a file into a struct successfully.
func Test_unmarshalConfigFile2(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
name: foobar
rack:
  name: r1
board:
  fromEnv: HOME
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &LocationConfig{}
	err := unmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "foobar", config.Name)
	assert.Equal(t, "r1", config.Rack.Name)
	assert.Equal(t, "", config.Rack.FromEnv)
	assert.Equal(t, "", config.Board.Name)
	assert.Equal(t, "HOME", config.Board.FromEnv)
}

// Test_unmarshalConfigFile3 tests unmarshalling data from a file into a struct when
// there is no data to unmarshal.
func Test_unmarshalConfigFile3(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", "", 0666)

	config := &LocationConfig{}
	err := unmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// Test_unmarshalConfigFile4 tests unmarshalling data from a file into a struct
// when none of the data in the file matches up with the struct.
func Test_unmarshalConfigFile4(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
foo: bar
version: 3
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &LocationConfig{}
	err := unmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// Test_unmarshalConfigFile5 tests unmarshalling data from a file into a struct
// when the file data is invalid yaml. This should result in an error.
func Test_unmarshalConfigFile5(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
name:: foo
rack
  name: bar
board:
  name: baz
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &LocationConfig{}
	err := unmarshalConfigFile(filename, config)
	assert.Error(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// Test_unmarshalConfigFile6 tests unmarshalling data from a file that doesn't exist.
func TestUnmarshalConfigFile6(t *testing.T) {
	config := &LocationConfig{}
	err := unmarshalConfigFile("/foo/bar/baz.yaml", config)
	assert.Error(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// Test_findConfigs_Env1 tests getting the filepaths for config files when the override
// environment variable is specified, but no configs are found in that path.
func Test_findConfigs_Env1(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_findConfigs_Env2 tests getting the filepaths for config files when the override
// environment variable is specified, and references a directory that doesn't exist.
func Test_findConfigs_Env2(t *testing.T) {
	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, "/a/b/c/d/")
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_findConfigs_Env3 tests getting the filepaths for config files when the override
// environment variable is specified, and references a file that doesn't exist.
func Test_findConfigs_Env3(t *testing.T) {
	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, "/a/b/c/foo.yaml")
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_findConfigs_Env4 tests getting the filepaths for config files when the override
// environment variable is specified, and references a file that exists and is a valid
// config type.
func Test_findConfigs_Env4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, foo)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_findConfigs_Env5 tests getting the filepaths for config files when the override
// environment variable is specified, and references a file that exists but is not a valid
// config type.
func Test_findConfigs_Env5(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.json", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, foo)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_findConfigs_Env6 tests getting the filepaths for config files when the override
// environment variable is specified, and references a directory that exists with a single
// valid file.
func Test_findConfigs_Env6(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_findConfigs_Env7 tests getting the filepaths for config files when the override
// environment variable is specified, and references a directory that exists with
// multiple valid files.
func Test_findConfigs_Env7(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add files to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", "", os.ModePerm)
	baz := test.WriteTempFile(t, "baz.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	paths, err := findConfigs(deviceConfigSearchPaths, EnvDeviceConfig, "")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(paths))
	assert.Equal(t, bar, paths[0])
	assert.Equal(t, baz, paths[1])
	assert.Equal(t, foo, paths[2])
}

// Test_findConfigs_Default1 tests getting the filepaths for config
// files when a search dir does not exist.
func Test_findConfigs_Default1(t *testing.T) {
	paths, err := findConfigs([]string{"/a/b/c/"}, "", "")
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_findConfigs_Default2 tests getting the filepaths for config
// files when a search dir is empty.
func Test_findConfigs_Default2(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	paths, err := findConfigs([]string{test.TempDir}, "", "")
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_findConfigs_Default3 tests getting the filepaths for config
// files when a search dir contains no valid configs.
func Test_findConfigs_Default3(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	_ = test.WriteTempFile(t, "foo.json", "", os.ModePerm)

	paths, err := findConfigs([]string{test.TempDir}, "", "")
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_findConfigs_Default4 tests getting the filepaths for config
// files when a search dir contains one valid config.
func Test_findConfigs_Default4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	paths, err := findConfigs([]string{test.TempDir}, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_findConfigs_Default5 tests getting the filepaths for config
// files when a search dir contains multiple valid configs.
func Test_findConfigs_Default5(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", "", os.ModePerm)
	baz := test.WriteTempFile(t, "baz.yaml", "", os.ModePerm)

	paths, err := findConfigs([]string{test.TempDir}, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(paths))
	assert.Equal(t, bar, paths[0])
	assert.Equal(t, baz, paths[1])
	assert.Equal(t, foo, paths[2])
}

// TestGetDeviceConfigsFromFile tests getting ConfigContext for all device configs found.
// In this case, no config files will be found.
func TestGetDeviceConfigsFromFile(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	ctxs, err := getDeviceConfigsFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

// TestGetDeviceConfigsFromFile2 tests getting ConfigContext for all device configs found.
// In this case, one config will be found, but will have invalid yaml.
func TestGetDeviceConfigsFromFile2(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
location::
rack
  name: foo
board:
  name: bar
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, foo)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	ctxs, err := getDeviceConfigsFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

// TestGetDeviceConfigsFromFile3 tests getting ConfigContext for all device configs found.
// In this case, one valid config is found.
func TestGetDeviceConfigsFromFile3(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	ctxs, err := getDeviceConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ctxs))
	assert.Equal(t, foo, ctxs[0].Source)

	assert.True(t, ctxs[0].IsDeviceConfig())
	cfg := ctxs[0].Config.(*DeviceConfig)
	assert.Equal(t, 3, cfg.Version)
}

// TestGetDeviceConfigsFromFile4 tests getting ConfigContext for all device configs found.
// In this case, multiple valid configs are found.
func TestGetDeviceConfigsFromFile4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDeviceConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvDeviceConfig)

	ctxs, err := getDeviceConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ctxs))

	barCtx := ctxs[0]
	assert.Equal(t, bar, barCtx.Source)
	assert.True(t, barCtx.IsDeviceConfig())
	cfg := barCtx.Config.(*DeviceConfig)
	assert.Equal(t, 3, cfg.Version)

	fooCtx := ctxs[1]
	assert.Equal(t, foo, fooCtx.Source)
	assert.True(t, fooCtx.IsDeviceConfig())
	cfg = fooCtx.Config.(*DeviceConfig)
	assert.Equal(t, 3, cfg.Version)
}

// TestGetPluginConfigFromFile tests getting the ConfigContext for the plugin config.
// In this case, no plugin config will be found.
func TestGetPluginConfigFromFile(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	ctx, err := getPluginConfigFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestGetPluginConfigFromFile2 tests getting the ConfigContext for the plugin config.
// In this case, the config will be found, but will have invalid yaml.
func TestGetPluginConfigFromFile2(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
location::
rack
  name: foo
board:
  name: bar
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "config.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, foo)
	defer test.RemoveEnv(t, EnvPluginConfig)

	ctx, err := getPluginConfigFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestGetPluginConfigFromFile3 tests getting the ConfigContext for the plugin config.
// In this case, one valid config is found.
func TestGetPluginConfigFromFile3(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
network:
  type: tcp
  address: "1.2.3.4:5001"
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "config.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	ctx, err := getPluginConfigFromFile()
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	assert.Equal(t, foo, ctx.Source)
	assert.True(t, ctx.IsPluginConfig())

	cfg := ctx.Config.(*PluginConfig)

	t.Logf("config: %#v", cfg)

	// check config values
	assert.Equal(t, 3, cfg.Version)
	assert.Equal(t, "tcp", cfg.Network.Type)
	assert.Equal(t, "1.2.3.4:5001", cfg.Network.Address)

	// check default values
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "serial", cfg.Settings.Mode)
	assert.Equal(t, true, cfg.Settings.Read.Enabled)
	assert.Equal(t, "1s", cfg.Settings.Read.Interval)
	assert.Equal(t, 100, cfg.Settings.Read.Buffer)
	assert.Equal(t, true, cfg.Settings.Write.Enabled)
	assert.Equal(t, "1s", cfg.Settings.Write.Interval)
	assert.Equal(t, 100, cfg.Settings.Write.Buffer)
	assert.Equal(t, 100, cfg.Settings.Write.Max)
	assert.Equal(t, "5m", cfg.Settings.Transaction.TTL)

	assert.Nil(t, cfg.Limiter)
}

// TestGetPluginConfigFromFile4 tests getting the ConfigContext for the plugin config.
// In this case, one valid config is found and some defaults are overridden.
func TestGetPluginConfigFromFile4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
debug: true
network:
  type: tcp
  address: "1.2.3.4:5001"
settings:
  mode: serial
  read:
    buffer: 150
    interval: 2s
    enabled: true
  write:
    buffer: 150
    max: 100
    interval: 2s
    enabled: false
limiter:
  rate: 100
  burst: 50
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "config.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	ctx, err := getPluginConfigFromFile()
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	assert.Equal(t, foo, ctx.Source)
	assert.True(t, ctx.IsPluginConfig())

	cfg := ctx.Config.(*PluginConfig)

	// check config values
	assert.Equal(t, 3, cfg.Version)
	assert.Equal(t, "tcp", cfg.Network.Type)
	assert.Equal(t, "1.2.3.4:5001", cfg.Network.Address)
	assert.Equal(t, 100, cfg.Limiter.Rate)
	assert.Equal(t, 50, cfg.Limiter.Burst)

	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, "serial", cfg.Settings.Mode)
	assert.Equal(t, true, cfg.Settings.Read.Enabled)
	assert.Equal(t, "2s", cfg.Settings.Read.Interval)
	assert.Equal(t, 150, cfg.Settings.Read.Buffer)
	assert.Equal(t, false, cfg.Settings.Write.Enabled)
	assert.Equal(t, "2s", cfg.Settings.Write.Interval)
	assert.Equal(t, 150, cfg.Settings.Write.Buffer)
	assert.Equal(t, 100, cfg.Settings.Write.Max)
	assert.Equal(t, "5m", cfg.Settings.Transaction.TTL)
}

// TestGetPluginConfigFromFile5 tests getting the ConfigContext for the plugin config.
// In this case, two possible configs will be found.
func TestGetPluginConfigFromFile5(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
locations:
- name: test
  rack:
    name: foo
  board:
    name: bar
`

	// Add files to the dir
	_ = test.WriteTempFile(t, "config.yaml", data, os.ModePerm)
	_ = test.WriteTempFile(t, "config.yml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	ctx, err := getPluginConfigFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctx)
}

// TestGetOutputTypeConfigsFromFile tests getting OutputType for all configs found.
// In this case, no config files will be found.
func TestGetOutputTypeConfigsFromFile(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvOutputTypeConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvOutputTypeConfig)

	ctxs, err := getOutputTypeConfigsFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

// TestGetOutputTypeConfigsFromFile2 tests getting OutputType for all configs found.
// In this case, one config will be found, but will have invalid yaml.
func TestGetOutputTypeConfigsFromFile2(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
location::
rack
  name: foo
board:
  name: bar
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvOutputTypeConfig, foo)
	defer test.RemoveEnv(t, EnvOutputTypeConfig)

	ctxs, err := getOutputTypeConfigsFromFile()
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

// TestGetOutputTypeConfigsFromFile3 tests getting OutputType for all configs found.
// In this case, one valid config is found.
func TestGetOutputTypeConfigsFromFile3(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvOutputTypeConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvOutputTypeConfig)

	ctxs, err := getOutputTypeConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ctxs))
	assert.Equal(t, foo, ctxs[0].Source)

	assert.True(t, ctxs[0].IsOutputTypeConfig())
	cfg := ctxs[0].Config.(*OutputType)
	assert.Equal(t, 3, cfg.Version)
}

// TestGetOutputTypeConfigsFromFile4 tests getting OutputType for all configs found.
// In this case, multiple valid configs are found.
func TestGetOutputTypeConfigsFromFile4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: 3
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvOutputTypeConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvOutputTypeConfig)

	ctxs, err := getOutputTypeConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ctxs))

	barCtx := ctxs[0]
	assert.Equal(t, bar, barCtx.Source)
	assert.True(t, barCtx.IsOutputTypeConfig())
	cfg := barCtx.Config.(*OutputType)
	assert.Equal(t, 3, cfg.Version)

	fooCtx := ctxs[1]
	assert.Equal(t, foo, fooCtx.Source)
	assert.True(t, fooCtx.IsOutputTypeConfig())
	cfg = fooCtx.Config.(*OutputType)
	assert.Equal(t, 3, cfg.Version)
}
