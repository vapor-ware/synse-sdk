package cfg

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/internal/test"
)

// TestIsValidConfig tests validating that a file is a potential config file.
func TestIsValidConfig(t *testing.T) {
	var testTable = []struct {
		desc    string
		isValid bool
		file    *test.FileInfo
	}{
		{
			desc:    "file is not valid -- is a directory",
			isValid: false,
			file:    test.NewFileInfo("test", os.ModeDir),
		},
		{
			desc:    "file is not valid -- is a json file",
			isValid: false,
			file:    test.NewFileInfo("test,json", os.ModePerm),
		},
		{
			desc:    "file is valid -- has .yml extension",
			isValid: true,
			file:    test.NewFileInfo("test.yml", os.ModePerm),
		},
		{
			desc:    "file is valid -- has .yaml extension",
			isValid: true,
			file:    test.NewFileInfo("test.yaml", os.ModePerm),
		},
	}

	for _, testCase := range testTable {
		result := isValidConfig(testCase.file)
		assert.Equal(t, testCase.isValid, result, testCase.desc)
	}
}

// Test_getConfigPathsFromDir_NoValidFiles tests getting the paths of the files that are
// valid config files from a directory. In this case, there are no valid configs.
func Test_getConfigPathsFromDir_NoValidFiles(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Test
	configs, err := getConfigPathsFromDir(test.TempDir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(configs))
}

// Test_getConfigPathsFromDir_NoDir tests getting the paths of the files that are
// valid config files from a directory when the directory doesn't exist.
func Test_getConfigPathsFromDir_NoDir(t *testing.T) {
	configs, err := getConfigPathsFromDir("a/b/c/d/e")
	assert.Error(t, err)
	assert.Nil(t, configs)
}

// Test_getConfigPathsFromDir_OneFile tests getting the paths of the files that are
// valid config files from a directory. In this case, there is only one valid config
// file, and some invalid.
func Test_getConfigPathsFromDir_OneFile(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	foo := test.WriteTempFile(t, "foo.yml", "", 0666)
	_ = test.WriteTempFile(t, "bar.json", "", 0666)

	// Test
	configs, err := getConfigPathsFromDir(test.TempDir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	assert.Equal(t, foo, configs[0])
}

// Test_getConfigPathsFromDir_MultipleFiles tests getting the paths of the files that are
// valid config files from a directory. In this case, is multiple valid config files.
func Test_getConfigPathsFromDir_MultipleFiles(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	foo := test.WriteTempFile(t, "foo.yml", "", 0666)
	bar := test.WriteTempFile(t, "bar.yaml", "", 0666)
	baz := test.WriteTempFile(t, "baz.yaml", "", 0666)

	// Test
	configs, err := getConfigPathsFromDir(test.TempDir)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(configs))
	assert.Equal(t, bar, configs[0])
	assert.Equal(t, baz, configs[1])
	assert.Equal(t, foo, configs[2])
}

// TestUnmarshalConfigFile tests unmarshalling data from a file into a struct successfully.
func TestUnmarshalConfigFile(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
version: 1.0
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &DeviceConfig{}
	err := UnmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "1.0", config.Version)
}

// TestUnmarshalConfigFile2 tests unmarshalling data from a file into a struct successfully.
func TestUnmarshalConfigFile2(t *testing.T) {
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

	config := &Location{}
	err := UnmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "foobar", config.Name)
	assert.Equal(t, "r1", config.Rack.Name)
	assert.Equal(t, "", config.Rack.FromEnv)
	assert.Equal(t, "", config.Board.Name)
	assert.Equal(t, "HOME", config.Board.FromEnv)
}

// TestUnmarshalConfigFile3 tests unmarshalling data from a file into a struct when
// there is no data to unmarshal.
func TestUnmarshalConfigFile3(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", "", 0666)

	config := &Location{}
	err := UnmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// TestUnmarshalConfigFile4 tests unmarshalling data from a file into a struct
// when none of the data in the file matches up with the struct.
func TestUnmarshalConfigFile4(t *testing.T) {
	// Set up a temporary directory for test data.
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data to write to file
	data := `
foo: bar
version: "1.0"
`

	// Add data to the temporary test directory
	filename := test.WriteTempFile(t, "foo.yml", data, 0666)

	config := &Location{}
	err := UnmarshalConfigFile(filename, config)
	assert.NoError(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// TestUnmarshalConfigFile5 tests unmarshalling data from a file into a struct
// when the file data is invalid yaml. This should result in an error.
func TestUnmarshalConfigFile5(t *testing.T) {
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

	config := &Location{}
	err := UnmarshalConfigFile(filename, config)
	assert.Error(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// TestUnmarshalConfigFile6 tests unmarshalling data from a file that doesn't exist.
func TestUnmarshalConfigFile6(t *testing.T) {
	config := &Location{}
	err := UnmarshalConfigFile("/foo/bar/baz.yaml", config)
	assert.Error(t, err)

	// Check that the config now has the correct fields.
	assert.Equal(t, "", config.Name)
	assert.Nil(t, config.Rack)
	assert.Nil(t, config.Board)
}

// Test_getDeviceConfigFilePaths_Env1 tests getting the filepaths for config files
// when the override environment variable is specified, but no configs are found
// in that path.
func Test_getDeviceConfigFilePaths_Env1(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_getDeviceConfigFilePaths_Env2 tests getting the filepaths for config files
// when the override environment variable is specified, and references a directory
// that doesn't exist.
func Test_getDeviceConfigFilePaths_Env2(t *testing.T) {
	// Set up the test env
	test.SetEnv(t, EnvDevicePath, "/a/b/c/d/")
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_getDeviceConfigFilePaths_Env3 tests getting the filepaths for config files
// when the override environment variable is specified, and references a file
// that doesn't exist.
func Test_getDeviceConfigFilePaths_Env3(t *testing.T) {
	// Set up the test env
	test.SetEnv(t, EnvDevicePath, "/a/b/c/foo.yaml")
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Equal(t, 0, len(paths))
}

// Test_getDeviceConfigFilePaths_Env4 tests getting the filepaths for config files
// when the override environment variable is specified, and references a file
// that exists and is a valid config type.
func Test_getDeviceConfigFilePaths_Env4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, foo)
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_getDeviceConfigFilePaths_Env5 tests getting the filepaths for config files
// when the override environment variable is specified, and references a file
// that exists but is not a valid config type.
func Test_getDeviceConfigFilePaths_Env5(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.json", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, foo)
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_getDeviceConfigFilePaths_Env6 tests getting the filepaths for config files
// when the override environment variable is specified, and references a directory
// that exists with a single valid file.
func Test_getDeviceConfigFilePaths_Env6(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_getDeviceConfigFilePaths_Env7 tests getting the filepaths for config files
// when the override environment variable is specified, and references a directory
// that exists with multiple valid files.
func Test_getDeviceConfigFilePaths_Env7(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add files to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", "", os.ModePerm)
	baz := test.WriteTempFile(t, "baz.yaml", "", os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	paths, err := getDeviceConfigFilePaths()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(paths))
	assert.Equal(t, bar, paths[0])
	assert.Equal(t, baz, paths[1])
	assert.Equal(t, foo, paths[2])
}

// Test_getDeviceConfigFilePaths_Default1 tests getting the filepaths for config
// files when a search dir does not exist.
func Test_getDeviceConfigFilePaths_Default1(t *testing.T) {
	// override the search paths for the test, use a bogus dir
	deviceConfigSearchPaths = []string{"/a/b/c/"}

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_getDeviceConfigFilePaths_Default2 tests getting the filepaths for config
// files when a search dir is empty.
func Test_getDeviceConfigFilePaths_Default2(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// override the search paths for the test, just use the temp dir
	deviceConfigSearchPaths = []string{test.TempDir}

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_getDeviceConfigFilePaths_Default3 tests getting the filepaths for config
// files when a search dir contains no valid configs.
func Test_getDeviceConfigFilePaths_Default3(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	_ = test.WriteTempFile(t, "foo.json", "", os.ModePerm)

	// override the search paths for the test, just use the temp dir
	deviceConfigSearchPaths = []string{test.TempDir}

	paths, err := getDeviceConfigFilePaths()
	assert.Error(t, err)
	assert.Nil(t, paths)
}

// Test_getDeviceConfigFilePaths_Default4 tests getting the filepaths for config
// files when a search dir contains one valid config.
func Test_getDeviceConfigFilePaths_Default4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)

	// override the search paths for the test, just use the temp dir
	deviceConfigSearchPaths = []string{test.TempDir}

	paths, err := getDeviceConfigFilePaths()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(paths))
	assert.Equal(t, foo, paths[0])
}

// Test_getDeviceConfigFilePaths_Default5 tests getting the filepaths for config
// files when a search dir contains multiple valid configs.
func Test_getDeviceConfigFilePaths_Default5(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", "", os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", "", os.ModePerm)
	baz := test.WriteTempFile(t, "baz.yaml", "", os.ModePerm)

	// override the search paths for the test, just use the temp dir
	deviceConfigSearchPaths = []string{test.TempDir}

	paths, err := getDeviceConfigFilePaths()
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
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	ctxs, err := GetDeviceConfigsFromFile()
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
	test.SetEnv(t, EnvDevicePath, foo)
	defer test.RemoveEnv(t, EnvDevicePath)

	ctxs, err := GetDeviceConfigsFromFile()
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
version: "1.0"
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	ctxs, err := GetDeviceConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ctxs))
	assert.Equal(t, foo, ctxs[0].Source)

	assert.True(t, ctxs[0].IsDeviceConfig())
	cfg := ctxs[0].Config.(*DeviceConfig)
	assert.Equal(t, "1.0", cfg.Version)
}

// TestGetDeviceConfigsFromFile4 tests getting ConfigContext for all device configs found.
// In this case, multiple valid configs are found.
func TestGetDeviceConfigsFromFile4(t *testing.T) {
	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	data := `
version: "1.0"
`

	// Add a file to the dir
	foo := test.WriteTempFile(t, "foo.yaml", data, os.ModePerm)
	bar := test.WriteTempFile(t, "bar.yml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvDevicePath, test.TempDir)
	defer test.RemoveEnv(t, EnvDevicePath)

	ctxs, err := GetDeviceConfigsFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ctxs))

	barCtx := ctxs[0]
	assert.Equal(t, bar, barCtx.Source)
	assert.True(t, barCtx.IsDeviceConfig())
	cfg := barCtx.Config.(*DeviceConfig)
	assert.Equal(t, "1.0", cfg.Version)

	fooCtx := ctxs[1]
	assert.Equal(t, foo, fooCtx.Source)
	assert.True(t, fooCtx.IsDeviceConfig())
	cfg = fooCtx.Config.(*DeviceConfig)
	assert.Equal(t, "1.0", cfg.Version)
}

// TestNewPluginConfig tests getting a new plugin config from file. In this case,
// we set the search path from env. The path will not exist.
func TestNewPluginConfig(t *testing.T) {
	// Reset the global viper instance for the test
	viper.Reset()

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, "foo/bar/baz")
	defer test.RemoveEnv(t, EnvPluginConfig)

	cfg, err := NewPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestNewPluginConfig2 tests getting a new plugin config from file. In this case,
// we set the search path from env. The path will exist, but will not contain any
// configs.
func TestNewPluginConfig2(t *testing.T) {
	// Reset the global viper instance for the test
	viper.Reset()

	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	cfg, err := NewPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestNewPluginConfig3 tests getting a new plugin config from file. In this case,
// we set the search path from env. The path will exist and will contain an invalid
// config.
func TestNewPluginConfig3(t *testing.T) {
	// Reset the global viper instance for the test
	viper.Reset()

	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data for the test config
	data := `
version:: "1.0"
debug true
settings
  mode: serial
`

	// Add a file to the dir
	_ = test.WriteTempFile(t, "config.yaml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	cfg, err := NewPluginConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestNewPluginConfig4 tests getting a new plugin config from file. In this case,
// we set the search path from env. The path will exist and contain a valid config.
func TestNewPluginConfig4(t *testing.T) {
	// Reset the global viper instance for the test
	viper.Reset()

	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data for the test config
	data := `
version: "1.0"
debug: true
network:
  type: tcp
  address: ":5001"
settings:
  mode: serial
  read:
    interval: 3s
`

	// Add a file to the dir
	_ = test.WriteTempFile(t, "config.yml", data, os.ModePerm)

	// Set up the test env
	test.SetEnv(t, EnvPluginConfig, test.TempDir)
	defer test.RemoveEnv(t, EnvPluginConfig)

	cfg, err := NewPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check the configured values
	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, "tcp", cfg.Network.Type)
	assert.Equal(t, ":5001", cfg.Network.Address)
	assert.Equal(t, "serial", cfg.Settings.Mode)
	assert.Equal(t, "3s", cfg.Settings.Read.Interval)

	// Check some of the default values that were not overwritten
	assert.Equal(t, 100, cfg.Settings.Read.Buffer)
	assert.Equal(t, true, cfg.Settings.Read.Enabled)
	assert.Equal(t, 100, cfg.Settings.Write.Buffer)
	assert.Equal(t, 100, cfg.Settings.Write.Max)
	assert.Equal(t, true, cfg.Settings.Write.Enabled)
	assert.Equal(t, "1s", cfg.Settings.Write.Interval)
	assert.Equal(t, "5m", cfg.Settings.Transaction.TTL)

	assert.Nil(t, cfg.Limiter)
}

// TestNewPluginConfig5 tests getting a new plugin config from file. In this case,
// we will use the default search path and find a valid config.
func TestNewPluginConfig5(t *testing.T) {
	// Reset the global viper instance for the test
	viper.Reset()

	// Set up the test dir
	test.SetupTestDir(t)
	defer test.ClearTestDir(t)

	// Data for the test config
	data := `
version: "1.0"
debug: false
network:
  type: unix
  address: "test.sock"
settings:
  mode: parallel
  read:
    interval: 1s
    buffer: 150
  write:
    enabled: false
limiter:
  rate: 100
  burst: 50
`

	// Add a file to the dir
	_ = test.WriteTempFile(t, "config.yml", data, os.ModePerm)

	// Set the search path to be the temp dir
	pluginConfigSearchPaths = []string{test.TempDir}

	cfg, err := NewPluginConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check the configured values
	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "unix", cfg.Network.Type)
	assert.Equal(t, "test.sock", cfg.Network.Address)
	assert.Equal(t, "parallel", cfg.Settings.Mode)
	assert.Equal(t, "1s", cfg.Settings.Read.Interval)
	assert.Equal(t, 150, cfg.Settings.Read.Buffer)
	assert.Equal(t, false, cfg.Settings.Write.Enabled)
	assert.Equal(t, 100, cfg.Limiter.Rate)
	assert.Equal(t, 50, cfg.Limiter.Burst)

	// Check some of the default values that were not overwritten
	assert.Equal(t, true, cfg.Settings.Read.Enabled)
	assert.Equal(t, 100, cfg.Settings.Write.Buffer)
	assert.Equal(t, 100, cfg.Settings.Write.Max)
	assert.Equal(t, "1s", cfg.Settings.Write.Interval)
	assert.Equal(t, "5m", cfg.Settings.Transaction.TTL)
}
