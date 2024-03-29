// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
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
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/v2/sdk/config"
	"github.com/vapor-ware/synse-sdk/v2/sdk/errors"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
	"github.com/vapor-ware/synse-sdk/v2/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

const (
	defaultWriteTimeout = 30 * time.Second
)

// Device is a single physical or virtual device which the Plugin manages.
//
// It defines all of the information known about the device, which typically
// comes from configuration file. A Device's supported actions are determined
// by the DeviceHandler which it is configured to use.
type Device struct {
	// Type is the type of the device. This is largely metadata that can
	// be used upstream to categorize the device.
	Type string

	// Info is a human-readable string that provides a summary of what
	// the device is or what it does.
	Info string

	// Tags is the set of tags that are associated with the device.
	Tags []*Tag

	// Data contains any plugin-specific configuration data for the device.
	Data map[string]interface{}

	// Context contains any contextual information which should be associated
	// with the device's reading(s).
	Context map[string]string

	// Handler is the name of the device's handler.
	Handler string

	// SortIndex is an optional 1-based index sort ordinal that is used upstream
	// (by Synse Server) to sort the device in Scan output. This can be set to
	// give devices custom ordering. If not set, they will be sorted based on
	// default sorting parameters.
	SortIndex int32

	// Alias is a human-readable alias for the device which can be used to
	// to reference it as well.
	Alias string

	// Transforms is a collection of the device reading transformers that are to
	// be applied to the device's reading(s). These are generally (but not limited)
	// to scale or convert the device reading(s). Typically, these will be defined
	// for general-purpose plugins where the device handler is not specific enough
	// to know the desired output of a device reading.
	//
	// Transformations are applied in the order in which they are defined.
	//
	// See the TransformConfig and Transformer godoc for more details.
	Transforms []Transformer

	// WriteTimeout defines the time within which a write action (transaction)
	// will remain valid for this device.
	WriteTimeout time.Duration

	// Output is the name of the Output that this device instance will use. This
	// is not needed for all devices/plugins, as many DeviceHandlers will already
	// know which output to use. This field is used in cases of generalized plugins,
	// such as Modbus-IP, where a generalized handler will need to map something
	// (like a set of registers) to a reading output.
	Output string

	// id is the unique ID for the device.
	id string

	// idName is the generated name of the device based on its components which
	// is ultimately used to generate its ID.
	idName string

	// handler is a pointer to the actual DeviceHandler for the device. This is
	// populated via the SDK on device loading and parsing and uses the Handler
	// field to match the name of the handler to the actual instance.
	handler *DeviceHandler
}

// NewDeviceFromConfig creates a new instance of a Device from its device prototype
// and device instance configuration. This is the primary and recommended way of
// building devices.
//
// These configuration components are loaded from config file.
//
// The device instance and prototype may specify a device handler name in their
// Handler field. This is only the name of the Handler, not a reference to the
// device Handler itself. A map[string]*DeviceHandler must be passed to this function
// to register the device handler with the device (in the private `handler` field,
// looking up the appropriate handler from the map using the `Handler` name.
//
// If no handler is found in the map for the device, an error is returned.
func NewDeviceFromConfig(
	proto *config.DeviceProto,
	instance *config.DeviceInstance,
	handlers map[string]*DeviceHandler,
) (*Device, error) {

	if proto == nil || instance == nil {
		return nil, fmt.Errorf("cannot create new device from nil config")
	}

	// Define variable for the Device fields that can be inherited from the
	// device prototype configuration.
	var (
		data         = map[string]interface{}{}
		context      = map[string]string{}
		tags         []string
		transforms   []Transformer
		handler      string
		deviceType   string
		writeTimeout time.Duration
	)

	// If inheritance is enabled, use the prototype defined value as the base. For
	// map and slice types, we need to make a copy so we do not mutate the prototype
	// values for other instances built off the same prototype.
	if !instance.DisableInheritance {
		for k, v := range proto.Data {
			data[k] = v
		}
		for k, v := range proto.Context {
			context[k] = v
		}
		tags = append(tags, proto.Tags...)
		handler = proto.Handler
		deviceType = proto.Type
		writeTimeout = proto.WriteTimeout

		for _, v := range proto.Transforms {
			t, err := NewTransformer(v)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"config": v,
				}).Error("[device] unable to create reading transformer")
				return nil, err
			}
			transforms = append(transforms, t)
		}
	}

	// If the instance also defines the same variable, we either need to merge
	// the values or overwrite them.

	// Merge instance data.
	if err := mergo.Map(&data, instance.Data, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
		log.WithField("error", err).Error("[device] failed merging device instance config: data")
		return nil, err
	}

	// Merge context data.
	if err := mergo.Map(&context, instance.Context, mergo.WithOverride); err != nil {
		log.WithField("error", err).Error("[device] failed merging device instance config: context")
	}

	// Merge tags. It is okay if the same tag is defined more than once, (e.g.
	// no need to error), but we do ultimately just want the set of tags.
	tags = append(tags, instance.Tags...)
	var deviceTags []*Tag
	encountered := map[string]struct{}{}
	for _, t := range tags {
		if _, ok := encountered[t]; !ok {
			encountered[t] = struct{}{}
			tag, err := NewTag(t)
			if err != nil {
				log.WithField("tag", t).Error("[device] failed to create new tag")
				return nil, err
			}
			deviceTags = append(deviceTags, tag)
		}
	}

	// Override type, if set.
	if instance.Type != "" {
		deviceType = instance.Type
	}

	// We require devices to have a type; error if there is none set.
	if deviceType == "" {
		log.WithFields(log.Fields{
			"prototype": proto,
			"instance":  instance,
		}).Error("[device] required field 'type' is missing")
		return nil, fmt.Errorf("new device: required field 'type' is missing")
	}

	// Override handler, if set.
	if instance.Handler != "" {
		handler = instance.Handler
	}

	// If no handler is set, fall back to using the type as the name
	// of the handler.
	if handler == "" {
		handler = deviceType
	}

	handlerFn, found := handlers[handler]
	if !found {
		return nil, fmt.Errorf("new device: unknown handler specified '%s'", handler)
	}

	// If an output is specified for the device, make sure that an output
	// with that name exists. If not, the device config is incorrect.
	if instance.Output != "" {
		if output.Get(instance.Output) == nil {
			log.WithFields(log.Fields{
				"prototype": proto,
				"instance":  instance,
			}).Error("[device] unknown output specified")
			return nil, fmt.Errorf("new device: unknown output specified '%s'", instance.Output)
		}
	}

	// Collect the instance transforms, then parse the transform configs
	// to validate they are correct.
	for _, v := range instance.Transforms {
		t, err := NewTransformer(v)
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err,
				"config": v,
			}).Error("[device] unable to create reading transformer")
			return nil, err
		}
		transforms = append(transforms, t)
	}

	// Override write timeout, if set.
	if instance.WriteTimeout != 0 {
		writeTimeout = instance.WriteTimeout
	}
	// Since we are merging proto + instance, we can't easily set a default value
	// in the config struct annotations, so make sure that the timeout is not 0 here.
	if writeTimeout == 0 {
		log.WithField("timeout", defaultWriteTimeout).Debug(
			"[device] no write timeout found in device config, using default timeout",
		)
		writeTimeout = defaultWriteTimeout
	}

	// Render any templates which may exist in the context.
	if err := parseContext(context); err != nil {
		return nil, err
	}

	d := &Device{
		Type:         deviceType,
		Tags:         deviceTags,
		Data:         data,
		Context:      context,
		Handler:      handler,
		Info:         instance.Info,
		SortIndex:    instance.SortIndex,
		Transforms:   transforms,
		WriteTimeout: writeTimeout,
		Output:       instance.Output,
		handler:      handlerFn,
	}

	if err := d.setAlias(instance.Alias); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"alias": instance.Alias,
		}).Error("[device] failed to set device alias")
		return nil, err
	}

	return d, nil
}

// parseContext is a utility function to parse the context map and render any templates
// which may exist in the context values.
func parseContext(ctx map[string]string) error {
	ctxTmpl := template.New("ctx").Funcs(template.FuncMap{
		"env": os.Getenv,
	})

	for k := range ctx {
		val := ctx[k]
		if val == "" {
			continue
		}

		tmpl, err := ctxTmpl.Parse(val)
		if err != nil {
			return err
		}
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, val); err != nil {
			return err
		}

		// NOTE: The SDK does not currently verify that the specified environment
		// variable was actually set. This is left to the configurer. Is this an
		// okay assumption/design choice?
		newVal := buf.String()
		if newVal == "" {
			log.WithFields(log.Fields{
				"key":   k,
				"value": val,
			}).Warn("[device] template detected in device context, but no value rendered for parsed template")
		}
		ctx[k] = newVal
	}
	return nil
}

// AliasContext is the context that is used to render alias templates.
type AliasContext struct {
	Meta   *PluginMetadata
	Device *Device
}

// setAlias sets the device alias for a given device.
func (device *Device) setAlias(conf *config.DeviceAlias) error {
	// If there is no DeviceAlias config, there is no alias set for the device.
	if conf == nil {
		return nil
	}

	// If the alias configuration specifies a name value, return that value.
	if conf.Name != "" {
		device.Alias = conf.Name
		return nil
	}

	// If the alias configuration specifies a template string, try and render the
	// template.
	if conf.Template != "" {
		ctx := &AliasContext{
			Meta:   &metadata,
			Device: device,
		}

		var buf bytes.Buffer

		t, err := template.New("alias").Funcs(template.FuncMap{
			"env": os.Getenv,
			"ctx": device.GetContext,
		}).Parse(conf.Template)
		if err != nil {
			return err
		}
		if err := t.Execute(&buf, ctx); err != nil {
			return err
		}

		device.Alias = buf.String()
		return nil
	}
	return nil
}

// GetContext gets a value out of the device's context map.
func (device *Device) GetContext(key string) string {
	return device.Context[key]
}

// GetHandler gets the DeviceHandler of the device.
func (device *Device) GetHandler() *DeviceHandler {
	return device.handler
}

// GetID gets the unique ID of the device.
func (device *Device) GetID() string {
	return device.id
}

// Read performs the read action for the device, as set by its DeviceHandler.
//
// If reading is not supported on the device, an UnsupportedCommandError is
// returned.
func (device *Device) Read() (*ReadContext, error) {
	if !device.IsReadable() {
		log.WithField("id", device.id).Debug("[device] device is not readable")
		return nil, &errors.UnsupportedCommandError{}
	}

	readings, err := device.handler.Read(device)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    device.id,
		}).Error("[device] failed to read from device")
		return nil, err
	}
	return NewReadContext(device, readings), nil
}

// Write performs the write action for the device, as set by its DeviceHandler.
//
// If writing is not supported on the device, an UnsupportedCommandError is
// returned.
func (device *Device) Write(data *WriteData) error {
	if !device.IsWritable() {
		log.WithField("id", device.id).Debug("[device] device is not writable")
		return &errors.UnsupportedCommandError{}
	}

	if len(device.handler.Actions) > 0 {
		hasAction := false
		for _, action := range device.handler.Actions {
			if data.Action == action {
				hasAction = true
			}
		}
		if !hasAction {
			return errors.InvalidArgumentErr(
				"unsupported write action '%v' for device %s", data.Action, device.id,
			)
		}
	}

	err := device.handler.Write(device, data)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    device.id,
		}).Error("[device] failed to write to device")

	}
	return err
}

// IsReadable checks if the Device is readable based on the presence/absence
// of a Read/BulkRead action defined in its DeviceHandler.
func (device *Device) IsReadable() bool {
	if device == nil {
		return false
	}
	return device.handler.CanRead() || device.handler.CanListen() || device.handler.CanBulkRead()
}

// IsWritable checks if the Device is writable based on the presence/absence
// of a Write action defined in its DeviceHandler.
func (device *Device) IsWritable() bool {
	if device == nil {
		return false
	}
	return device.handler.CanWrite()
}

// encode translates the Device to the corresponding gRPC Device message.
func (device *Device) encode() *synse.V3Device {
	var tags = make([]*synse.V3Tag, len(device.Tags))
	for i, t := range device.Tags {
		tags[i] = t.Encode()
	}

	// If the device is writable, include the pre-defined write actions.
	var actions []string
	if device.IsWritable() {
		actions = device.handler.Actions
	}

	// outputs are augmented into this in server.go, prior to it being returned
	// as a gRPC response.
	return &synse.V3Device{
		Timestamp: utils.GetCurrentTime(),
		Id:        device.id,
		Type:      device.Type,
		Info:      device.Info,
		Alias:     device.Alias,
		Metadata:  device.Context,
		SortIndex: device.SortIndex,
		Tags:      tags,
		Capabilities: &synse.V3DeviceCapability{
			Mode: device.handler.GetCapabilitiesMode(),
			Write: &synse.V3WriteCapability{
				Actions: actions,
			},
		},
	}
}
