package sdk

type pluginAction func(p *Plugin) error
type deviceAction func(p *Plugin, d *Device) error
