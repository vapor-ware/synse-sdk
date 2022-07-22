package sdk

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/funcs"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// Errors relating to Transformer creation and application.
var (
	ErrNilTransformConfig = errors.New("cannot create transformer: nil config")
	ErrUnknownTransformFn = errors.New("unknown transform apply function specified")
)

// A Transformer is something which transforms a device's raw reading value(s).
// This transformation is typically a scaling or type conversion. This is generally
// used more for general-purpose plugins where the device handler is not specific
// to the device/output and the reading value may need to be coerced into a desired
// output.
type Transformer interface {

	// Apply the transformation to the given reading value.
	Apply(reading *output.Reading) error

	// Get the name of the transformer.
	Name() string
}

// ApplyTransformer is a device reading Transformer which applies pre-defined
// functions to a device's reading(s).
type ApplyTransformer struct {
	Func *funcs.Func
}

// NewApplyTransformer creates a new device reading Transformer which is used
// to apply pre-defined functions to a device's reading(s). The SDK has some
// built-in functions in the 'funcs' package. A plugin may also register its
// own. Functions are referenced by name.
func NewApplyTransformer(fn string) (*ApplyTransformer, error) {
	f := funcs.Get(fn)
	if f == nil {
		log.WithFields(log.Fields{
			"fn": fn,
		}).Error("[transform] unknown transform function specified")
		return nil, ErrUnknownTransformFn
	}

	return &ApplyTransformer{
		Func: f,
	}, nil
}

// Apply the transformer function to the given reading value.
func (t *ApplyTransformer) Apply(reading *output.Reading) error {
	val, err := t.Func.Call(reading.Value)
	if err != nil {
		return err
	}
	reading.Value = val
	return nil
}

// Name returns a human-readable name for the apply transformer.
func (t *ApplyTransformer) Name() string {
	return fmt.Sprintf("apply [%v]", t.Func.Name)
}

// ScaleTransformer is a device reading transformer which scales a device's
// reading(s) by a multiplicative factor.
type ScaleTransformer struct {
	Factor float64
}

// NewScaleTransformer creates a new device reading Transformer which is used
// to scale readings by a multiplicative factor.
//
// The scaling factor is specified as a string, but should resolve to a numeric.
// By  default, it will have a value of 1 (e.g. no-op). Negative values and
// fractional values are supported. This can be the value itself, e.g. "0.01",
// or a mathematical representation of the value, e.g. "1e-2". Dividing is the same
// as multiplying by a fraction (e.g. "/ 2" == "* 0.5").
func NewScaleTransformer(factor string) (*ScaleTransformer, error) {
	var scaleBy float64 = 1
	var err error

	if factor != "" {
		scaleBy, err = strconv.ParseFloat(factor, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"factor": factor,
				"error":  err,
			}).Error("[transform] failed to create scale transformer: bad factor")
			return nil, err
		}
	}

	return &ScaleTransformer{
		Factor: scaleBy,
	}, nil
}

// Apply the transformer scaling factor to the given reading value.
func (t *ScaleTransformer) Apply(reading *output.Reading) error {
	if t.Factor == 1 {
		// Nothing to scale.
		return nil
	}
	return reading.Scale(t.Factor)
}

// Name returns a human-readable name for the scale transformer.
func (t *ScaleTransformer) Name() string {
	return fmt.Sprintf("scale [%v]", t.Factor)
}

// NewTransformer creates a new device reading Transformer from the provided
// TransformConfig. If the configuration is incorrect or specifies unsupported
// values, an error is returned.
func NewTransformer(cfg *config.TransformConfig) (Transformer, error) {
	if cfg == nil {
		return nil, ErrNilTransformConfig
	}

	log.WithFields(log.Fields{
		"apply": cfg.Apply,
		"scale": cfg.Scale,
	}).Debug("[transform] creating new device reading transformer")

	// Verify the config is valid and does not contain multiple operations.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if cfg.Apply != "" {
		return NewApplyTransformer(cfg.Apply)
	} else if cfg.Scale != "" {
		return NewScaleTransformer(cfg.Scale)
	} else {
		return nil, errors.New("no transformer operation configured")
	}
}
