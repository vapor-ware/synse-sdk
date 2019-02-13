package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// fixme (etd): add logging throughout. will do this once this gets folded back into the
//  main sdk folder and has access to the sdk logger.

const (
	// Yaml-extension configuration files.
	ExtYaml = "yaml"
)

// validExts maps the file extension name constant to all supported
// extensions for that format.
var validExts = map[string][]string{
	ExtYaml: {".yml", ".yaml"},
}

// Loader is used to load configurations from file(s) and environment and unify
// them all into a singular configuration.
//
// While other configuration solutions exist, such as spf13/viper, micro/go-config,
// etc..., they lack good support for loading and unifying multiple configuration
// files, as is needed here since Device configs can be specified across any
// number of files. Additionally, many of the existing popular solutions contain
// many features that Synse does not need, which introduces *a lot* of dependency
// bloat.
//
// This configuration loader is meant to be a simple file & environment only
// solution. This could be made into its own package, external to the SDK in
// the future.
type Loader struct {
	// Name is the name of the config being loaded. This is optional and is only used
	// when logging messages.
	Name string

	// SearchPaths defines the paths to search for configuration files. These paths
	// are searched for configs in the order in which they are defined until config(s)
	// are found.
	SearchPaths []string

	// Ext is the file extension format of the config files.
	Ext string

	// EnvOverride defines the environment variable which can be used to override
	// the search paths/file name. This env variable will be ignored when searching
	// for variables starting with EnvPrefix, so this is free to also use the
	// EnvPrefix.
	EnvOverride string

	// EnvPrefix is the prefix for configuration environment variables.
	EnvPrefix string

	// FileName is the name of the file to use. If this is set, only this file will
	// be loaded. If this is not set, all files with the specified extension in
	// the specified search paths will be loaded. This can be specified with or
	// without a file extension.
	FileName string

	// The files which were found to match the loader parameters on search.
	// This is populated by the `search()` function and used in the `read()`
	// function.
	files []string

	// The contents of the files found to match the loader parameters. This is
	// populated by the `loadEnv()` and `read()` functions and is used in the
	// `merge()` function.
	data []map[string]interface{}

	// The merged config contents. This is populated by the `merge()` function.
	merged map[string]interface{}
}

// NewYamlLoader creates a new loader which is configured to read YAML configuration
// file(s).
func NewYamlLoader(name string) *Loader {
	return &Loader{
		Name: name,
		Ext:  ExtYaml,
	}
}

// AddSearchPaths adds search paths to the config Loader.
//
// These paths are searched in the order that they are defined.
func (loader *Loader) AddSearchPaths(paths ...string) {
	loader.SearchPaths = append(loader.SearchPaths, paths...)
}

// Load loads the configuration based on the Loader's configurations. The loading
// process consists of:
//
// 1. Checking for environment overrides
// 2. Searching for the specified config files, if any
// 3. Reading in any found config files
// 4. Loading any environmental configuration
// 5. Merging all found configurations together
//
// Environmental configuration takes precedence, so it will override any values
// that were set in config files.
func (loader *Loader) Load() error {
	if err := loader.checkOverrides(); err != nil {
		return err
	}

	if err := loader.search(); err != nil {
		return err
	}

	if err := loader.read(); err != nil {
		return err
	}

	if err := loader.loadEnv(); err != nil {
		return err
	}

	if err := loader.merge(); err != nil {
		return err
	}

	return nil
}

// Scan the merged configuration values into the provided go type. The type
// passed to Scan should be a pointer to a zero-value struct, e.g.
//
//    config := &Config{}
//    loader.Scan(config)
//
func (loader *Loader) Scan(out interface{}) error {
	if loader.merged == nil || len(loader.merged) == 0 {
		// fixme
		return fmt.Errorf("unable to scan, no merged content, did you Load first")
	}
	return mapstructure.WeakDecode(loader.merged, out)
}

// checkOverrides checks to see if an override configuration file/path is set
// in the environment, and if so, updates the loader to use those values.
func (loader *Loader) checkOverrides() error {
	// If there is no environment override, there is nothing to do here.
	if loader.EnvOverride == "" {
		return nil
	}

	value := os.Getenv(loader.EnvOverride)

	// If there is no value set for the environment override, then the
	// config is not overridden and we should continue searching as normal.
	if value == "" {
		return nil
	}

	// Get info on the specified path. We will need to know whether it is
	// a specific file (load that file), or a directory (load all configs in
	// that directory).
	info, err := os.Stat(value)
	if err != nil {
		return err
	}

	// If a directory is specified, we will use it as the only search path.
	// If the config was set via env, we expect it to be there, so we do not
	// want to fall back to the default search paths.
	if info.IsDir() {
		loader.SearchPaths = []string{value}
		loader.FileName = ""
	} else {
		dir, file := filepath.Split(value)

		if !loader.isValidExt(file) {
			// fixme: error handling
			return fmt.Errorf("env override specified invalid file extension")
		}

		// The specified file matches the expected extension, so we can use it
		// as our expected file.
		loader.SearchPaths = []string{dir}
		loader.FileName = file
	}
	return nil
}

// loadEnv searches the environment for variables which start with the specified
// EnvPrefix. All found variables are collected and transformed into a data map.
func (loader *Loader) loadEnv() error {
	// Search for configuration environment variables. Exclude the EnvOverride
	// variable, if it is set.
	if loader.EnvPrefix != "" {
		envConfig := make(map[string]interface{})

		for _, env := range os.Environ() {
			if strings.HasPrefix(env, loader.EnvPrefix) {
				pair := strings.SplitN(env, "=", 2)

				// If the key matches the environment override key, ignore it.
				if pair[0] == loader.EnvOverride {
					continue
				}

				// Get the (possibly nested) keys, excluding the EnvPrefix.
				keys := strings.Split(strings.ToLower(pair[0]), "_")[1:]
				value := pair[1]

				// To build the potentially nested config from env, reverse
				// the keys and build the map from the most inner item, working
				// outwards.
				for i := len(keys)/2 - 1; i >= 0; i-- {
					opp := len(keys) - 1 - i
					keys[i], keys[opp] = keys[opp], keys[i]
				}

				tmp := make(map[string]interface{})
				for idx, key := range keys {
					if idx == 0 {
						tmp[key] = value
						continue
					}
					tmp = map[string]interface{}{key: tmp}
				}

				if err := mergo.Map(&envConfig, tmp); err != nil {
					return err
				}
			}
		}

		if len(envConfig) > 0 {
			loader.data = append(loader.data, envConfig)
		}
	}
	return nil
}

// search searches for configuration files based on the specified search
// path(s) and file name given to the Loader.
func (loader *Loader) search() error {
	for _, path := range loader.SearchPaths {
		dirContents, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, file := range dirContents {
			if loader.isValidFile(file) {
				fileName := filepath.Join(path, file.Name())
				loader.files = append(loader.files, fileName)
			}
		}
	}
	return nil
}

// read reads each of the found configuration files into a data mapping.
// These data mappings are collected by the Loader to be merged later.
func (loader *Loader) read() error {
	for _, path := range loader.files {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		res := map[string]interface{}{}

		switch loader.Ext {
		case ExtYaml:
			err := yaml.Unmarshal(data, &res)
			if err != nil {
				return err
			}
			loader.data = append(loader.data, res)
		default:
			// fixme
			return fmt.Errorf("unsupported file format: %v", loader.Ext)
		}
	}
	return nil
}

// merge merges all of the data mappings from all config files and environment
// variables that were found, generating a single unified config.
func (loader *Loader) merge() error {
	for _, data := range loader.data {
		// If there are any nil maps, there is nothing to merge.
		if data == nil {
			continue
		}

		// If there are any empty maps, there is nothing to merge.
		if len(data) == 0 {
			continue
		}

		// Merge the data map.
		if err := mergo.Map(&loader.merged, data, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			return err
		}
	}
	return nil
}

// isValidFile checks whether a given file is valid by seeing whether it meets the
// constraints set by the config Loader.
func (loader *Loader) isValidFile(info os.FileInfo) bool {
	if !info.IsDir() {
		// If a FileName was specified, check that the file matches that name.
		if loader.FileName != "" {
			name := info.Name()
			ext := filepath.Ext(name)

			// If the loader's filename does not have an extension, check against the
			// fileName without its extension.
			if filepath.Ext(loader.FileName) == "" {
				name = strings.TrimRight(info.Name(), ext)
			}

			// Check that the names match.
			if loader.FileName != name {
				return false
			}
		}

		// Check that the file extension is supported.
		return loader.isValidExt(info.Name())
	}
	return false
}

// isValidExt checks whether a given path has a supported extension for the Loader.
func (loader *Loader) isValidExt(path string) bool {
	exts, ok := validExts[loader.Ext]
	if !ok {
		// fixme: log error
		return false
	}

	ext := filepath.Ext(path)
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}
