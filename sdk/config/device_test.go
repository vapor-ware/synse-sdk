package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformConfig_Validate_Ok(t *testing.T) {
	tests := []struct {
		name string
		cfg  TransformConfig
	}{
		{
			name: "empty config",
			cfg:  TransformConfig{},
		},
		{
			name: "only apply",
			cfg: TransformConfig{
				Apply: "testing",
			},
		},
		{
			name: "only scale",
			cfg: TransformConfig{
				Scale: "testing",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.cfg.Validate()
			assert.NoError(t, err, test.name)
		})
	}
}

func TestTransformConfig_Validate_Error(t *testing.T) {
	cfg := TransformConfig{
		Apply: "testing",
		Scale: "testing",
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTransform, err)
}
