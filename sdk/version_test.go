package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEmptyVersionInfo tests creating a VersionInfo populated with
// its default 'empty' values.
func TestEmptyVersionInfo(t *testing.T) {
	v := emptyVersionInfo()

	assert.Equal(t, "-", v.VersionString)
	assert.Equal(t, "-", v.GoVersion)
	assert.Equal(t, "-", v.GitTag)
	assert.Equal(t, "-", v.GitCommit)
	assert.Equal(t, "-", v.BuildDate)
}

// TestVersionInfo_Merge tests merging two VersionInfo instances
// with fields that conflict.
func TestVersionInfo_Merge(t *testing.T) {
	v1 := VersionInfo{
		BuildDate:     "yesterday",
		GitCommit:     "123",
		GitTag:        "git-tag-2",
		GoVersion:     "go1.8",
		VersionString: "2",
	}

	v2 := VersionInfo{
		BuildDate:     "today",
		GitCommit:     "abc",
		GitTag:        "git-tag-1",
		GoVersion:     "go1.9",
		VersionString: "1",
	}

	// Merge v2 into v1
	v1.Merge(&v2)

	// All fields from v2 should be taken, so v1 and v2 should now be equal
	assert.Equal(t, v2, v1)
}

// TestVersionInfo_Merge2 tests merging two VersionInfo instances
// with fields that do not conflict.
func TestVersionInfo_Merge2(t *testing.T) {
	v1 := VersionInfo{
		BuildDate: "today",
		GitTag:    "tag1",
	}

	v2 := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
	}

	expected := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
		GitTag:        "tag1",
		BuildDate:     "today",
	}

	// Merge v2 into v1
	v1.Merge(&v2)
	assert.Equal(t, expected, v1)
}

// TestVersionInfo_Merge3 tests merging two VersionInfo instances
// where one VersionInfo is the empty default.
func TestVersionInfo_Merge3(t *testing.T) {
	v1 := emptyVersionInfo()
	v2 := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
	}

	expected := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
		GitTag:        "-",
		BuildDate:     "-",
		GoVersion:     "-",
	}

	// Merge v2 into v1
	v1.Merge(&v2)
	assert.Equal(t, expected, *v1)
}

// TestVersionInfo_Merge4 tests merging two VersionInfo instances
// where some fields conflict, but others do not.
func TestVersionInfo_Merge4(t *testing.T) {
	v1 := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
		GitTag:        "1.1",
		BuildDate:     "1-2-3",
		GoVersion:     "1.8",
	}
	v2 := VersionInfo{
		VersionString: "2",
		GitCommit:     "def",
		GitTag:        "1.2",
	}

	expected := VersionInfo{
		VersionString: "2",
		GitCommit:     "def",
		GitTag:        "1.2",
		BuildDate:     "1-2-3",
		GoVersion:     "1.8",
	}

	// Merge v2 into v1
	v1.Merge(&v2)
	assert.Equal(t, expected, v1)
}
