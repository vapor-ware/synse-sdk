package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetPluginMeta tests setting the global plugin meta-information.
func TestSetPluginMeta(t *testing.T) {
	// Make sure it is empty to begin with
	assert.IsType(t, meta{}, metainfo)
	assert.Equal(t, "", metainfo.Name)
	assert.Equal(t, "", metainfo.Maintainer)
	assert.Equal(t, "", metainfo.Description)
	assert.Equal(t, "", metainfo.VCS)

	// Set the metainfo
	SetPluginMeta("name", "maintainer", "desc", "vcs")

	// Check that it has changed
	assert.IsType(t, meta{}, metainfo)
	assert.Equal(t, "name", metainfo.Name)
	assert.Equal(t, "maintainer", metainfo.Maintainer)
	assert.Equal(t, "desc", metainfo.Description)
	assert.Equal(t, "vcs", metainfo.VCS)
}
