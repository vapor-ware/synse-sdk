// Synse SDK
// Copyright (c) 2019-2020 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	sdkError "github.com/vapor-ware/synse-sdk/sdk/errors"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
	"github.com/vapor-ware/synse-sdk/sdk/utils"
	"gopkg.in/yaml.v2"
)

// Definitions of supported configuration file types.
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

	// The policy used for the most recent configuration Load.
	policy policy.Policy

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
//
// This function takes a policy as a parameter. The policy determines whether the
// configuration file is required or not. In cases where it is required and not
// found, Load will return an error. If the config is optional and not found, no
// error will be returned.
func (loader *Loader) Load(pol policy.Policy) error {
	log.WithFields(log.Fields{
		"loader": loader.Name,
		"paths":  loader.SearchPaths,
		"name":   loader.FileName,
		"ext":    loader.Ext,
		"policy": pol,
	}).Info("[config] loading configuration")

	loader.policy = pol

	if err := loader.checkOverrides(); err != nil {
		return err
	}

	if err := loader.search(pol); err != nil {
		return err
	}

	if err := loader.read(pol); err != nil {
		return err
	}

	if err := loader.loadEnv(); err != nil {
		return err
	}

	if err := loader.merge(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"loader": loader.Name,
		"paths":  loader.SearchPaths,
		"name":   loader.FileName,
		"ext":    loader.Ext,
		"policy": loader.policy,
		"data":   utils.RedactPasswords(loader.merged),
	}).Info("[config] successfully loaded configuration data")
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
		if loader.policy == policy.Optional {
			log.WithFields(log.Fields{
				"loader": loader.Name,
				"policy": loader.policy,
			}).Debug("[config] no config found for Scan")
			return nil
		}

		log.WithFields(log.Fields{
			"loader": loader.Name,
			"policy": loader.policy,
		}).Error("[config] unable to scan config: no merged configs found, but are required")
		return errors.New("config: no merged config to scan")
	}

	log.WithFields(log.Fields{
		"type": reflect.TypeOf(out),
	}).Debug("[config] scanning config into struct")

	if err := defaults.Set(out); err != nil {
		log.WithField("error", err).Error("[config] failed to set config defaults")
		return err
	}

	// Use a custom decoder config for decoding so we can pick up time durations.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           out,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		log.WithField("error", err).Error("[config] failed to decode config data")
		return err
	}

	return decoder.Decode(loader.merged)
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

	log.Debug("[config] loading ENV overrides")

	// Get info on the specified path. We will need to know whether it is
	// a specific file (load that file), or a directory (load all configs in
	// that directory).
	info, err := os.Stat(value)
	if err != nil {
		log.WithFields(log.Fields{
			"path": value,
		}).Error("[config] failed to stat path")
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
			log.WithFields(log.Fields{
				"value": value,
			}).Error("[config] invalid file extension from ENV")
			return errors.New("config: invalid file extension")
		}

		// The specified file matches the expected extension, so we can use it
		// as our expected file.
		loader.SearchPaths = []string{dir}
		loader.FileName = file
	}
	log.WithFields(log.Fields{
		"file":  loader.FileName,
		"paths": loader.SearchPaths,
	}).Info("[config] ENV overrides loaded")

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
				log.WithField("env", env).Debug("[config] found prefixed ENV variable")
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

				if len(tmp) != 0 {
					log.WithFields(log.Fields{
						"data": tmp,
					}).Debug("[config] loaded environment data")
				}

				if err := mergo.Map(&envConfig, tmp); err != nil {
					log.WithField("error", err).Error("[config] failed to merge env config")
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
func (loader *Loader) search(pol policy.Policy) error {
	var required = pol == policy.Required

	for _, path := range loader.SearchPaths {
		plog := log.WithFields(log.Fields{
			"loader": loader.Name,
			"path":   path,
			"policy": pol,
		})

		dirContents, err := ioutil.ReadDir(path)
		if err != nil {
			plog.Debug("[config] no config found, searching next path")
			continue
		}

		var foundInPath bool
		for _, file := range dirContents {
			if loader.isValidFile(file) {
				plog.WithFields(log.Fields{
					"file": file.Name(),
				}).Info("[config] found matching config")
				foundInPath = true
				fileName := filepath.Join(path, file.Name())
				loader.files = append(loader.files, fileName)
			}
		}

		// If configuration was found in the current path, break to stop
		// searching. We do not want to search all potential config paths
		// if we have already found matching config.
		if foundInPath {
			break
		}
	}

	// If the config is required, make sure that we found something. If no
	// config was found on any of the search paths, return an error.
	if required && len(loader.files) == 0 {
		log.Error("[config] config is required but not found")
		return sdkError.NewConfigsNotFoundError(loader.SearchPaths)
	}

	return nil
}

// read reads each of the found configuration files into a data mapping.
// These data mappings are collected by the Loader to be merged later.
func (loader *Loader) read(pol policy.Policy) error {
	if pol == policy.Required && len(loader.files) == 0 {
		log.WithFields(log.Fields{
			"policy": pol,
			"files":  loader.files,
		}).Error("[config] no files loaded")
		return sdkError.NewConfigsNotFoundError(loader.SearchPaths)
	}

	for _, path := range loader.files {
		log.WithField("file", path).Info("[config] reading config file")
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.WithField("error", err).Error("[config] failed to read file")
			return err
		}

		res := map[string]interface{}{}

		switch loader.Ext {
		case ExtYaml:
			err := yaml.Unmarshal(data, &res)
			if err != nil {
				log.WithField("error", err).Error("[config] failed to unmarshal config data")
				return err
			}
			log.WithFields(log.Fields{
				"file": path,
				"data": utils.RedactPasswords(res),
			}).Debug("[config] loaded configuration from file")
			loader.data = append(loader.data, res)
		default:
			log.WithField("ext", loader.Ext).Error("[config] unsupported file format")
			return fmt.Errorf("config: unsupported file format '%v'", loader.Ext)
		}
	}
	return nil
}

// merge merges all of the data mappings from all config files and environment
// variables that were found, generating a single unified config.
func (loader *Loader) merge() error {
	log.Debug("[config] merging configuration sources")
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
			log.Error("[config] failed to merge config data")
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
		log.WithField("ext", loader.Ext).Debug("[config] file extension not supported")
		return false
	}

	ext := filepath.Ext(path)
	for _, e := range exts {
		if e == ext {
			return true
		}
	}

	log.WithField("path", path).Debug("[config] path contains unsupported extension")
	return false
}
