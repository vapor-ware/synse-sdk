package config

import (
	"os"
	"testing"

	"github.com/vapor-ware/synse-sdk/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestNewYamlLoader(t *testing.T) {
	cases := []struct {
		name string
	}{
		{""},
		{"1"},
		{"test"},
		{"longer name description"},
	}

	for _, c := range cases {
		loader := NewYamlLoader(c.name)

		assert.Equal(t, c.name, loader.Name)
		assert.Equal(t, ExtYaml, loader.Ext)

		assert.Empty(t, loader.files)
		assert.Empty(t, loader.data)
		assert.Empty(t, loader.merged)
	}
}

func TestLoader_AddSearchPaths(t *testing.T) {
	cases := []struct {
		paths []string
	}{
		{},
		{[]string{"foo"}},
		{[]string{"foo/bar", "1/2/3"}},
		{[]string{"foo/bar", "1/2/3", "./../..."}},
	}

	for _, c := range cases {
		loader := Loader{}
		assert.Empty(t, loader.SearchPaths)

		// Load the search paths
		loader.AddSearchPaths(c.paths...)
		assert.Equal(t, c.paths, loader.SearchPaths)

		// Load the search paths again - this should append to the
		// already existing search paths.
		loader.AddSearchPaths(c.paths...)
		assert.Equal(t, len(c.paths)*2, len(loader.SearchPaths))
	}
}

func TestLoader_Scan_Ok(t *testing.T) {
	type TestData struct {
		Value string
	}
	td := &TestData{}

	// Create a loader and manually set the merged data.
	loader := Loader{}
	loader.merged = map[string]interface{}{
		"value": "foo",
	}

	// Scan the config into the TestData struct.
	err := loader.Scan(td)
	assert.NoError(t, err)
	assert.Equal(t, "foo", td.Value)
}

func TestLoader_Scan_ConfigDoesNotMatch(t *testing.T) {
	type TestData struct {
		Value string
	}
	td := &TestData{}

	// Create a loader and manually set the merged data.
	loader := Loader{}
	loader.merged = map[string]interface{}{
		"no_match": "foo",
	}

	// Scan the config into the TestData struct.
	err := loader.Scan(td)
	assert.NoError(t, err)
	assert.Equal(t, "", td.Value)
}

func TestLoader_Scan_NoMergedData(t *testing.T) {
	type TestData struct {
		Value string
	}
	td := &TestData{}

	// Create a loader, do not specify any merged data.
	loader := Loader{}

	// Scan the config into the TestData struct.
	err := loader.Scan(td)
	assert.Error(t, err)
	assert.Equal(t, &TestData{}, td)
}

func TestLoader_checkOverrides_noOverride(t *testing.T) {
	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
	}

	err := loader.checkOverrides()
	assert.NoError(t, err)
	assert.Equal(t, []string{"placeholder"}, loader.SearchPaths)
	assert.Equal(t, "placeholder", loader.FileName)
}

func TestLoader_checkOverrides_overrideNotSet(t *testing.T) {
	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
		EnvOverride: "SDKTEST_OVERRIDE",
	}

	err := loader.checkOverrides()
	assert.NoError(t, err)
	assert.Equal(t, []string{"placeholder"}, loader.SearchPaths)
	assert.Equal(t, "placeholder", loader.FileName)
}

func TestLoader_checkOverrides_overrideNotExist(t *testing.T) {
	overrideEnv := "SDKTEST_OVERRIDE"
	assert.NoError(t, os.Setenv(overrideEnv, "./testdata/nonexistent-file.yaml"))
	defer func() {
		assert.NoError(t, os.Unsetenv(overrideEnv))
	}()

	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
		EnvOverride: overrideEnv,
	}

	err := loader.checkOverrides()
	assert.Error(t, err)
	assert.Equal(t, []string{"placeholder"}, loader.SearchPaths)
	assert.Equal(t, "placeholder", loader.FileName)
}

func TestLoader_checkOverrides_overrideExists_dir(t *testing.T) {
	overrideEnv := "SDKTEST_OVERRIDE"
	assert.NoError(t, os.Setenv(overrideEnv, "./testdata"))
	defer func() {
		assert.NoError(t, os.Unsetenv(overrideEnv))
	}()

	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
		EnvOverride: overrideEnv,
	}

	err := loader.checkOverrides()
	assert.NoError(t, err)
	assert.Equal(t, []string{"./testdata"}, loader.SearchPaths)
	assert.Equal(t, "", loader.FileName)
}

func TestLoader_checkOverrides_overrideExists_file(t *testing.T) {
	overrideEnv := "SDKTEST_OVERRIDE"
	assert.NoError(t, os.Setenv(overrideEnv, "./testdata/test.yaml"))
	defer func() {
		assert.NoError(t, os.Unsetenv(overrideEnv))
	}()

	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
		EnvOverride: overrideEnv,
	}

	err := loader.checkOverrides()
	assert.NoError(t, err)
	assert.Equal(t, []string{"./testdata/"}, loader.SearchPaths)
	assert.Equal(t, "test.yaml", loader.FileName)
}

func TestLoader_checkOverrides_overrideExists_invalidFile(t *testing.T) {
	overrideEnv := "SDKTEST_OVERRIDE"
	assert.NoError(t, os.Setenv(overrideEnv, "./testdata/test.json"))
	defer func() {
		assert.NoError(t, os.Unsetenv(overrideEnv))
	}()

	loader := Loader{
		FileName:    "placeholder",
		SearchPaths: []string{"placeholder"},
		Ext:         ExtYaml,
		EnvOverride: overrideEnv,
	}

	err := loader.checkOverrides()
	assert.Error(t, err)
	assert.Equal(t, []string{"placeholder"}, loader.SearchPaths)
	assert.Equal(t, "placeholder", loader.FileName)
}

func TestLoader_loadEnv_noEnvPrefix(t *testing.T) {
	loader := Loader{}
	assert.Empty(t, loader.data)

	err := loader.loadEnv()
	assert.NoError(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_loadEnv_okSingleEnv(t *testing.T) {
	assert.NoError(t, os.Setenv("SDKTEST_FOO", "2"))
	defer func() {
		assert.NoError(t, os.Unsetenv("SDKTEST_FOO"))
	}()

	loader := Loader{
		EnvPrefix: "SDKTEST",
	}
	assert.Empty(t, loader.data)

	err := loader.loadEnv()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.data))

	expected := map[string]interface{}{
		"foo": "2",
	}
	assert.Equal(t, expected, loader.data[0])
}

func TestLoader_loadEnv_okMultipleEnv(t *testing.T) {
	assert.NoError(t, os.Setenv("SDKTEST_FOO", "2"))
	assert.NoError(t, os.Setenv("SDKTEST_ABC_DEF", "test"))
	assert.NoError(t, os.Setenv("SDKTEST_ABC_XYZ", "value"))
	defer func() {
		assert.NoError(t, os.Unsetenv("SDKTEST_FOO"))
		assert.NoError(t, os.Unsetenv("SDKTEST_ABC_DEF"))
		assert.NoError(t, os.Unsetenv("SDKTEST_ABC_XYZ"))
	}()

	loader := Loader{
		EnvPrefix: "SDKTEST",
	}
	assert.Empty(t, loader.data)

	err := loader.loadEnv()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.data))

	expected := map[string]interface{}{
		"foo": "2",
		"abc": map[string]interface{}{
			"def": "test",
			"xyz": "value",
		},
	}
	assert.Equal(t, expected, loader.data[0])
}

func TestLoader_loadEnv_withEnvOverride(t *testing.T) {
	assert.NoError(t, os.Setenv("SDKTEST_OVERRIDE", "/tmp"))
	assert.NoError(t, os.Setenv("SDKTEST_FOO", "2"))
	defer func() {
		assert.NoError(t, os.Unsetenv("SDKTEST_OVERRIDE"))
		assert.NoError(t, os.Unsetenv("SDKTEST_FOO"))
	}()

	loader := Loader{
		EnvOverride: "SDKTEST_OVERRIDE",
		EnvPrefix:   "SDKTEST",
	}
	assert.Empty(t, loader.data)

	err := loader.loadEnv()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.data))

	expected := map[string]interface{}{
		"foo": "2",
	}
	assert.Equal(t, expected, loader.data[0])
}

func TestLoader_loadEnv_noPrefixesMatch(t *testing.T) {
	loader := Loader{
		EnvPrefix: "SDKTEST",
	}
	assert.Empty(t, loader.data)

	err := loader.loadEnv()
	assert.NoError(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_search_ok(t *testing.T) {
	loader := Loader{
		Ext:         ExtYaml,
		SearchPaths: []string{"./testdata"},
	}
	assert.Empty(t, loader.data)

	err := loader.search()
	assert.NoError(t, err)
	// we should only match two files in testdata/: invalid.yaml, test.yaml
	assert.Equal(t, 2, len(loader.files))
}

func TestLoader_search_ok2(t *testing.T) {
	loader := Loader{
		Ext:         ExtYaml,
		SearchPaths: []string{"."}, // no valid configs in current dir
	}
	assert.Empty(t, loader.data)

	err := loader.search()
	assert.NoError(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_search_noPaths(t *testing.T) {
	loader := Loader{
		Ext:         ExtYaml,
		SearchPaths: []string{},
	}
	assert.Empty(t, loader.data)

	err := loader.search()
	assert.NoError(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_read_ok(t *testing.T) {
	loader := Loader{
		Ext: ExtYaml,
	}
	loader.files = []string{"./testdata/test.yaml"}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.data))

	expected := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	assert.Equal(t, expected, loader.data[0])
}

func TestLoader_read_ok2(t *testing.T) {
	loader := Loader{
		Ext: ExtYaml,
	}
	loader.files = []string{
		"./testdata/test.yaml",
		"./testdata/test.yaml",
	}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(loader.data))

	expected := map[string]interface{}{
		"foo": 1,
		"bar": 2,
	}
	assert.Equal(t, expected, loader.data[0])
	assert.Equal(t, expected, loader.data[1])
}

func TestLoader_read_noFiles(t *testing.T) {
	loader := Loader{
		Ext: ExtYaml,
	}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.NoError(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_read_fileNotFound(t *testing.T) {
	loader := Loader{
		Ext: ExtYaml,
	}
	loader.files = []string{"./testdata/nonexistent-file.yaml"}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.Error(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_read_badExt(t *testing.T) {
	loader := Loader{
		Ext: "invalidExt",
	}
	loader.files = []string{"./testdata/test.yaml"}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.Error(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_read_badData(t *testing.T) {
	loader := Loader{
		Ext: ExtYaml,
	}
	loader.files = []string{"./testdata/invalid.yaml"}
	assert.Empty(t, loader.data)

	err := loader.read()
	assert.Error(t, err)
	assert.Empty(t, loader.data)
}

func TestLoader_merge_nilData(t *testing.T) {
	loader := Loader{}
	loader.data = nil

	err := loader.merge()
	assert.NoError(t, err)
	assert.Nil(t, loader.merged)
}

func TestLoader_merge_emptyData(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{}

	err := loader.merge()
	assert.NoError(t, err)
	assert.Nil(t, loader.merged)
}

func TestLoader_merge_nilDataset(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{nil}

	err := loader.merge()
	assert.NoError(t, err)
	assert.Nil(t, loader.merged)
}

func TestLoader_merge_emptyDataset(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{{}}

	err := loader.merge()
	assert.NoError(t, err)
	assert.Nil(t, loader.merged)
}

func TestLoader_merge_singleMap(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{
		{
			"foo":    "bar",
			"a":      1,
			"values": []int{1, 2},
		},
	}

	err := loader.merge()
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"foo":    "bar",
		"a":      1,
		"values": []int{1, 2},
	}
	assert.Equal(t, expected, loader.merged)
}

func TestLoader_merge_multiMapNoConflict(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{
		{
			"foo":    "bar",
			"a":      1,
			"values": []int{1, 2},
		},
		{
			"bar": "baz",
			"abc": "xyz",
		},
	}

	err := loader.merge()
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"foo":    "bar",
		"a":      1,
		"values": []int{1, 2},
		"bar":    "baz",
		"abc":    "xyz",
	}
	assert.Equal(t, expected, loader.merged)
}

func TestLoader_merge_multiMapOverride(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{
		{
			"foo":    "bar",
			"a":      1,
			"values": []int{1, 2},
		},
		{
			"foo": "bar",
			"a":   2,
		},
	}

	err := loader.merge()
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"foo":    "bar",
		"a":      2,
		"values": []int{1, 2},
	}
	assert.Equal(t, expected, loader.merged)
}

func TestLoader_merge_multiMapSliceAppend(t *testing.T) {
	loader := Loader{}
	loader.data = []map[string]interface{}{
		{
			"foo":    "bar",
			"a":      1,
			"values": []int{1, 2},
		},
		{
			"foo":    "bar",
			"a":      1,
			"values": []int{3, 4},
		},
	}

	err := loader.merge()
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"foo":    "bar",
		"a":      1,
		"values": []int{1, 2, 3, 4},
	}
	assert.Equal(t, expected, loader.merged)
}

func TestLoader_isValidFile(t *testing.T) {
	cases := []struct {
		fileName string
		ext      string
		info     os.FileInfo
		expected bool
	}{
		{
			// File is a directory.
			fileName: "foo",
			ext:      "",
			info:     test.NewFileInfo("foo", os.ModeDir),
			expected: false,
		},
		{
			// File is a directory.
			fileName: "foo/bar",
			ext:      "",
			info:     test.NewFileInfo("foo/bar", os.ModeDir),
			expected: false,
		},
		{
			// FileName not specified, ext matches file.
			fileName: "",
			ext:      "yaml",
			info:     test.NewFileInfo("foo.yml", os.ModePerm),
			expected: true,
		},
		{
			// FileName not specified, ext does not match file.
			fileName: "",
			ext:      "yaml",
			info:     test.NewFileInfo("foo.json", os.ModePerm),
			expected: false,
		},
		{
			// FileName specified with extension, matches file.
			fileName: "config.yaml",
			ext:      "yaml",
			info:     test.NewFileInfo("config.yaml", os.ModePerm),
			expected: true,
		},
		{
			// FileName specified with extension, does not match file.
			fileName: "config.yaml",
			ext:      "yaml",
			info:     test.NewFileInfo("other", os.ModePerm),
			expected: false,
		},
		{
			// FileName specified without extension, matches file.
			fileName: "config",
			ext:      "yaml",
			info:     test.NewFileInfo("config.yaml", os.ModePerm),
			expected: true,
		},
		{
			// FileName specified without extension, does not match file.
			fileName: "config",
			ext:      "yaml",
			info:     test.NewFileInfo("other", os.ModePerm),
			expected: false,
		},
	}

	for i, c := range cases {
		loader := Loader{
			Ext:      c.ext,
			FileName: c.fileName,
		}

		actual := loader.isValidFile(c.info)
		assert.Equal(t, c.expected, actual, "case: %d", i)
	}
}

func TestLoader_isValidExt(t *testing.T) {
	cases := []struct {
		ext      string
		path     string
		expected bool
	}{
		{
			// YAML extension for YAML file.
			ext:      "yaml",
			path:     "/foo/bar.yml",
			expected: true,
		},
		{
			// YAML extension for YAML file.
			ext:      "yaml",
			path:     "/foo/bar.yaml",
			expected: true,
		},
		{
			// YAML extension for YAML file.
			ext:      "yaml",
			path:     "bar.yml",
			expected: true,
		},
		{
			// YAML extension for non-YAML file.
			ext:      "yaml",
			path:     "/foo/bar.json",
			expected: false,
		},
		{
			// YAML extension for file with no extension.
			ext:      "yaml",
			path:     "/foo/bar",
			expected: false,
		},
		{
			// Unknown extension.
			ext:      "foo",
			path:     "/foo/bar.yml",
			expected: false,
		},
	}

	for i, c := range cases {
		loader := Loader{
			Ext: c.ext,
		}

		actual := loader.isValidExt(c.path)
		assert.Equal(t, c.expected, actual, "case %d", i)
	}
}

// -----
// Testing SDK use cases (device config, plugin config)
// -----
// TODO (): add more tests cases here..

func TestLoader_Load(t *testing.T) {
	l := NewYamlLoader("test")
	l.AddSearchPaths("./testdata/device")

	err := l.Load()
	assert.NoError(t, err)

	d := &Devices{}
	err = l.Scan(d)
	assert.NoError(t, err)
	assert.Equal(t, 3, d.Version)
	assert.Equal(t, 2, len(d.Devices))
}

type Tst struct {
	Foo int
	Bar int
}

func TestLoader_Load2(t *testing.T) {
	err := os.Setenv("SDKTEST_FOO", "1")
	assert.NoError(t, err)

	err = os.Setenv("SDKTEST_BAR", "2")
	assert.NoError(t, err)

	defer func() {
		err = os.Unsetenv("SDKTEST_FOO")
		assert.NoError(t, err)
		err = os.Unsetenv("SDKTEST_BAR")
		assert.NoError(t, err)
	}()

	l := NewYamlLoader("test")
	l.EnvPrefix = "SDKTEST"

	err = l.Load()
	assert.NoError(t, err)

	d := &Tst{}
	err = l.Scan(d)
	assert.NoError(t, err)

	assert.Equal(t, 1, d.Foo)
	assert.Equal(t, 2, d.Bar)
}
