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

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
)

// Test creating a new ID when a nil config is passed in.
func TestNewPluginID_nilConf(t *testing.T) {
	meta := PluginMetadata{}

	id, err := newPluginID(nil, &meta)
	assert.Error(t, err)
	assert.Nil(t, id)
}

// Test creating a new ID when a nil metadata is passed in.
func TestNewPluginID_nilMeta(t *testing.T) {
	conf := config.IDSettings{}

	id, err := newPluginID(&conf, nil)
	assert.Error(t, err)
	assert.Nil(t, id)
}

// Test creating a new ID when the config and meta are empty.
func TestNewPluginID_empty(t *testing.T) {
	conf := config.IDSettings{}
	meta := PluginMetadata{}

	id, err := newPluginID(&conf, &meta)
	assert.Error(t, err)
	assert.Nil(t, id)
}

// TODO: mock machine ID for testing
func TestNewPluginID_useMachineID(t *testing.T) {
	//conf := config.IDSettings{
	//	UseMachineID: true,
	//	UsePluginTag: false,
	//}
	//meta := PluginMetadata{
	//	Name: "foo",
	//	Maintainer: "bar",
	//}
	//
	//id, err := newPluginID(&conf, &meta)
	//assert.NoError(t, err)
	//
	//assert.Equal(t, "", id.name)
	//assert.Equal(t, "", id.uuid.String())
}

func TestNewPluginID_useTag(t *testing.T) {
	conf := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: true,
		UseEnv:       []string{},
		UseCustom:    []string{},
	}
	meta := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id, err := newPluginID(&conf, &meta)
	assert.NoError(t, err)

	assert.Equal(t, "bar/foo", id.name)
	assert.Equal(t, "1d916ec2-f015-5f3e-869d-36ef30dce23f", id.uuid.String())
}

func TestNewPluginID_useEnv(t *testing.T) {
	envOne := "TEST_VAL_ONE"
	envTwo := "TEST_VAL_TWO"

	assert.NoError(t, os.Setenv(envOne, "one"))
	assert.NoError(t, os.Setenv(envTwo, "two"))
	defer func() {
		assert.NoError(t, os.Unsetenv(envOne))
		assert.NoError(t, os.Unsetenv(envTwo))
	}()

	conf := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: false,
		UseEnv: []string{
			envOne,
			envTwo,
		},
		UseCustom: []string{},
	}
	meta := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id, err := newPluginID(&conf, &meta)
	assert.NoError(t, err)

	assert.Equal(t, "one.two", id.name)
	assert.Equal(t, "62444264-4604-5b06-840a-3e5ab9848c46", id.uuid.String())
}

func TestNewPluginID_useCustom(t *testing.T) {
	conf := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: false,
		UseEnv:       []string{},
		UseCustom: []string{
			"a", "b", "c", "d",
		},
	}
	meta := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id, err := newPluginID(&conf, &meta)
	assert.NoError(t, err)

	assert.Equal(t, "a.b.c.d", id.name)
	assert.Equal(t, "b3ddc32c-b165-5230-801a-66a1138a3942", id.uuid.String())
}

func TestNewPluginID_useMultiple(t *testing.T) {
	envOne := "TEST_VAL_ONE"
	envTwo := "TEST_VAL_TWO"

	assert.NoError(t, os.Setenv(envOne, "one"))
	assert.NoError(t, os.Setenv(envTwo, "two"))
	defer func() {
		assert.NoError(t, os.Unsetenv(envOne))
		assert.NoError(t, os.Unsetenv(envTwo))
	}()

	conf := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: true,
		UseEnv: []string{
			envOne,
			envTwo,
		},
		UseCustom: []string{
			"a", "b", "c",
		},
	}
	meta := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id, err := newPluginID(&conf, &meta)
	assert.NoError(t, err)

	assert.Equal(t, "bar/foo.one.two.a.b.c", id.name)
	assert.Equal(t, "b87c4dd2-0017-54cc-9cc4-56fcef7991a5", id.uuid.String())
}

func TestNewPluginID_checkEquality(t *testing.T) {
	// Generate the first ID.
	conf1 := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: true,
		UseEnv:       []string{},
		UseCustom: []string{
			"1", "2", "3",
		},
	}
	meta1 := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id1, err := newPluginID(&conf1, &meta1)
	assert.NoError(t, err)

	// Generate the second ID.
	conf2 := config.IDSettings{
		UseMachineID: false,
		UsePluginTag: true,
		UseEnv:       []string{},
		UseCustom: []string{
			"1", "2", "3",
		},
	}
	meta2 := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id2, err := newPluginID(&conf2, &meta2)
	assert.NoError(t, err)

	// Verify that the two IDs are the same.
	assert.Equal(t, id1.uuid.String(), id2.uuid.String())
	assert.Equal(t, id1.name, id2.name)
}

func TestPluginID_NewNamespacedID(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{
			name: "",
			id:   "806d8a6c-710a-505f-a8d5-51d2a5baf710",
		},
		{
			name: "foo",
			id:   "3356863c-1adf-5d27-a6c2-8ab41cf816d0",
		},
		{
			name: "foo",
			id:   "3356863c-1adf-5d27-a6c2-8ab41cf816d0",
		},
		{
			name: "bar",
			id:   "d5a734ec-e9e7-5bc7-91da-b25e149a919c",
		},
		{
			name: "1234567890",
			id:   "1a200720-ff49-5af9-a1cd-58f692957d6d",
		},
	}

	// Create the base ID
	conf := config.IDSettings{
		UsePluginTag: true,
	}
	meta := PluginMetadata{
		Name:       "foo",
		Maintainer: "bar",
	}

	id, err := newPluginID(&conf, &meta)
	assert.NoError(t, err)

	for i, c := range cases {
		subID := id.NewNamespacedID(c.name)
		assert.Equal(t, c.id, subID, "case: %d", i)
	}
}
