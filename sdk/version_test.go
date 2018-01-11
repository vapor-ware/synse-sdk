package sdk

import (
	"testing"
)

func TestEmptyVersionInfo(t *testing.T) {
	v := emptyVersionInfo()
	if v.VersionString != "-" {
		t.Errorf("VersionInfo.VersionString should be '-'")
	}
	if v.GoVersion != "-" {
		t.Errorf("VersionInfo.GoVersion should be '-'")
	}
	if v.GitTag != "-" {
		t.Errorf("VersionInfo.GitTag should be '-'")
	}
	if v.GitCommit != "-" {
		t.Errorf("VersionInfo.GitCommit should be '-'")
	}
	if v.BuildDate != "-" {
		t.Errorf("VersionInfo.BuildDate should be '-'")
	}
}

func TestVersionInfo_Merge(t *testing.T) {
	v1 := VersionInfo{}
	v2 := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
	}

	expected := VersionInfo{
		VersionString: "1",
		GitCommit:     "abc",
	}

	v1.Merge(&v2)
	if v1 != expected {
		t.Errorf("VersionInfo.Merge(%#v) => %#v, want %#v", v2, v1, expected)
	}
}

func TestVersionInfo_Merge2(t *testing.T) {
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

	v1.Merge(&v2)
	if *v1 != expected {
		t.Errorf("VersionInfo.Merge(%#v) => %#v, want %#v", v2, *v1, expected)
	}
}

func TestVersionInfo_Merge3(t *testing.T) {
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

	v1.Merge(&v2)
	if v1 != expected {
		t.Errorf("VersionInfo.Merge(%#v) => %#v, want %#v", v2, v1, expected)
	}
}
