package cfg

import (
	"testing"

	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/assert"
)

type testFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	sys     interface{}
}

// NOTE: this is not the correct way to check the isDir bit, but since
// we will only set it this way for tests, this is fine here.
func (f testFileInfo) IsDir() bool        { return f.mode == os.ModeDir }
func (f testFileInfo) ModTime() time.Time { return f.modTime }
func (f testFileInfo) Mode() os.FileMode  { return f.mode }
func (f testFileInfo) Name() string       { return f.name }
func (f testFileInfo) Size() int64        { return f.size }
func (f testFileInfo) Sys() interface{}   { return f.sys }

// TestIsValidConfig tests validating that a file is a potential config file.
func TestIsValidConfig(t *testing.T) {
	var testTable = []struct {
		desc    string
		isValid bool
		file    testFileInfo
	}{
		{
			desc:    "file is not valid -- is a directory",
			isValid: false,
			file: testFileInfo{
				name: "test",
				mode: os.ModeDir,
			},
		},
		{
			desc:    "file is not valid -- is a json file",
			isValid: false,
			file: testFileInfo{
				name: "test.json",
			},
		},
		{
			desc:    "file is valid -- has .yml extension",
			isValid: true,
			file: testFileInfo{
				name: "test.yml",
			},
		},
		{
			desc:    "file is valid -- has .yaml extension",
			isValid: true,
			file: testFileInfo{
				name: "test.yaml",
			},
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
	// Set up a temp directory for testing.
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(dir)
		if err != nil {
			t.Error(err)
		}
	}()

	// Test
	configs, err := getConfigPathsFromDir(dir)
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
	// Set up a temp directory for testing.
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(dir)
		if err != nil {
			t.Error(err)
		}
	}()

	// Add a valid config and an invalid config file to the dir
	if err = ioutil.WriteFile(filepath.Join(dir, "foo.yml"), []byte{}, 0666); err != nil {
		t.Error(err)
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "bar.json"), []byte{}, 0666); err != nil {
		t.Error(err)
	}

	// Test
	configs, err := getConfigPathsFromDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	assert.Equal(t, filepath.Join(dir, "foo.yml"), configs[0])
}

// Test_getConfigPathsFromDir_MultipleFiles tests getting the paths of the files that are
// valid config files from a directory. In this case, is multiple valid config files.
func Test_getConfigPathsFromDir_MultipleFiles(t *testing.T) {
	// Set up a temp directory for testing.
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(dir)
		if err != nil {
			t.Error(err)
		}
	}()

	// Add a valid config and an invalid config file to the dir
	if err = ioutil.WriteFile(filepath.Join(dir, "foo.yml"), []byte{}, 0666); err != nil {
		t.Error(err)
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "bar.yaml"), []byte{}, 0666); err != nil {
		t.Error(err)
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "baz.yaml"), []byte{}, 0666); err != nil {
		t.Error(err)
	}

	// Test
	configs, err := getConfigPathsFromDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(configs))
	assert.Equal(t, filepath.Join(dir, "bar.yaml"), configs[0])
	assert.Equal(t, filepath.Join(dir, "baz.yaml"), configs[1])
	assert.Equal(t, filepath.Join(dir, "foo.yml"), configs[2])
}
