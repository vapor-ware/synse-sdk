package sdk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/funcs"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func TestNewApplyTransformer(t *testing.T) {
	transformer, err := NewApplyTransformer("FtoC")
	assert.NoError(t, err)
	assert.NotNil(t, transformer)
	assert.Equal(t, &funcs.FtoC, transformer.Func)
}

func TestNewApplyTransformer_Error(t *testing.T) {
	transformer, err := NewApplyTransformer("does not exist")
	assert.Error(t, err)
	assert.Equal(t, ErrUnknownTransformFn, err)
	assert.Nil(t, transformer)
}

func TestApplyTransformer_Apply(t *testing.T) {
	transformer := ApplyTransformer{
		Func: &funcs.Func{
			Name: "test func",
			Fn: func(value interface{}) (interface{}, error) {
				return value.(int) * 2, nil
			},
		},
	}

	reading := output.Reading{Value: 2}
	err := transformer.Apply(&reading)
	assert.NoError(t, err)
	assert.Equal(t, 4, reading.Value)
}

func TestApplyTransformer_Apply_Error(t *testing.T) {
	transformer := ApplyTransformer{
		Func: &funcs.Func{
			Name: "test func",
			Fn: func(value interface{}) (interface{}, error) {
				return nil, fmt.Errorf("err")
			},
		},
	}

	reading := output.Reading{Value: 2}
	err := transformer.Apply(&reading)
	assert.Error(t, err)
	assert.Equal(t, 2, reading.Value)
}

func TestApplyTransformer_Name(t *testing.T) {
	transformer := ApplyTransformer{
		Func: &funcs.Func{
			Name: "test func",
			Fn: func(value interface{}) (interface{}, error) {
				return nil, fmt.Errorf("err")
			},
		},
	}

	assert.Equal(t, "apply [test func]", transformer.Name())
}

func TestNewScaleTransformer(t *testing.T) {
	transformer, err := NewScaleTransformer("2")
	assert.NoError(t, err)
	assert.NotNil(t, transformer)
	assert.Equal(t, float64(2), transformer.Factor)
}

func TestNewScaleTransformer_Error(t *testing.T) {
	transformer, err := NewScaleTransformer("not a factor")
	assert.Error(t, err)
	assert.Nil(t, transformer)
}

func TestScaleTransformer_Apply(t *testing.T) {
	transformer := ScaleTransformer{
		Factor: 2,
	}

	reading := output.Reading{Value: 2}
	err := transformer.Apply(&reading)
	assert.NoError(t, err)
	assert.Equal(t, float64(4), reading.Value)
}

func TestScaleTransformer_Apply_NoOp(t *testing.T) {
	transformer := ScaleTransformer{
		Factor: 1,
	}

	reading := output.Reading{Value: 2}
	err := transformer.Apply(&reading)
	assert.NoError(t, err)
	assert.Equal(t, 2, reading.Value)
}

func TestScaleTransformer_Apply_Error(t *testing.T) {
	transformer := ScaleTransformer{
		Factor: 0,
	}

	reading := output.Reading{Value: 2}
	err := transformer.Apply(&reading)
	assert.Error(t, err)
	assert.Equal(t, 2, reading.Value)
}

func TestScaleTransformer_Name(t *testing.T) {
	transformer := ScaleTransformer{
		Factor: 3,
	}

	assert.Equal(t, "scale [3]", transformer.Name())
}

func TestNewTransformer_NilConfig(t *testing.T) {
	transformer, err := NewTransformer(nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilTransformConfig, err)
	assert.Nil(t, transformer)
}

func TestNewTransformer_InvalidConfig(t *testing.T) {
	transformer, err := NewTransformer(&config.TransformConfig{
		Scale: "2",
		Apply: "FtoC",
	})
	assert.Error(t, err)
	assert.Nil(t, transformer)
}

func TestNewTransformer_Scale(t *testing.T) {
	transformer, err := NewTransformer(&config.TransformConfig{
		Scale: "3",
	})
	assert.NoError(t, err)
	assert.NotNil(t, transformer)
	assert.Equal(t, "scale [3]", transformer.Name())
}

func TestNewTransformer_Apply(t *testing.T) {
	transformer, err := NewTransformer(&config.TransformConfig{
		Apply: "FtoC",
	})
	assert.NoError(t, err)
	assert.NotNil(t, transformer)
	assert.Equal(t, "apply [FtoC]", transformer.Name())
}

func TestNewTransformer_NoTransforms(t *testing.T) {
	transformer, err := NewTransformer(&config.TransformConfig{})
	assert.Error(t, err)
	assert.Nil(t, transformer)
}
