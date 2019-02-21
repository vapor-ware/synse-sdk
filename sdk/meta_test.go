// Synse SDK
// Copyright (c) 2019 Vapor IO
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

package sdk

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//// TestSetPluginMeta tests setting the global plugin meta-information.
//func TestSetPluginMeta(t *testing.T) {
//	// Make sure it is empty to begin with
//	assert.IsType(t, meta{}, metainfo)
//	assert.Equal(t, "", metainfo.Name)
//	assert.Equal(t, "", metainfo.Maintainer)
//	assert.Equal(t, "", metainfo.Description)
//	assert.Equal(t, "", metainfo.VCS)
//
//	// Set the metainfo
//	SetPluginMeta("name", "maintainer", "desc", "vcs")
//
//	// Check that it has changed
//	assert.IsType(t, meta{}, metainfo)
//	assert.Equal(t, "name", metainfo.Name)
//	assert.Equal(t, "maintainer", metainfo.Maintainer)
//	assert.Equal(t, "desc", metainfo.Description)
//	assert.Equal(t, "vcs", metainfo.VCS)
//}
//
//// TestMetaLog tests logging out the metadata.
//func TestMetaLog(t *testing.T) {
//	metainfo.log()
//}
//
//// Test_makeTag tests making metainfo tags.
//func Test_makeTag(t *testing.T) {
//	var testTable = []struct {
//		name       string
//		maintainer string
//		expected   string
//	}{
//		{
//			name:       "test",
//			maintainer: "vapor io",
//			expected:   "vapor-io/test",
//		},
//		{
//			name:       "Test",
//			maintainer: "vaporio",
//			expected:   "vaporio/test",
//		},
//		{
//			name:       "Simple Plugin",
//			maintainer: "Vapor I-0",
//			expected:   "vapor-i_0/simple-plugin",
//		},
//		{
//			name:       "Simple Modbus-over-IP",
//			maintainer: "Vapor IO",
//			expected:   "vapor-io/simple-modbus_over_ip",
//		},
//		{
//			name:       "99 bottles of beer",
//			maintainer: "The Wall",
//			expected:   "the-wall/99-bottles-of-beer",
//		},
//	}
//
//	for _, testCase := range testTable {
//		actual := makeTag(testCase.name, testCase.maintainer)
//		assert.Equal(t, testCase.expected, actual)
//	}
//}
