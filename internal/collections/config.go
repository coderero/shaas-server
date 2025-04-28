package collections

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	ConfigCollectionName        = "config"
	ClimateConfigCollectionName = "climate_config"
	LDRConfigCollectionName     = "ldr_config"
	MotionConfigCollectionName  = "motion_config"
	RelayConfigCollectionName   = "relay_config"
)

type Config struct {
	ID      string `json:"id"`
	Device  string `json:"device_id"`
	Version int    `json:"version"`
}

func (*Config) Name() string {
	return ConfigCollectionName
}

func (*Config) Schema() *core.Collection {
	collection := core.NewBaseCollection(ConfigCollectionName, ConfigCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = device.user.id &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  DevicesCollectionName,
			Name:          "device",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     1,
		},
		&core.NumberField{
			Name:     "version",
			Required: true,
		},
	)

	return collection
}

type ClimateConfig struct {
	ID         string `json:"id"`
	SensorId   string `json:"sensor_id"`
	Lable      string `json:"lable"`
	Config     string `json:"config"`
	Dht22Port  int    `json:"dht22_port"`
	AQIPort    int    `json:"aqi_port"`
	HasBuzzer  bool   `json:"has_buzzer"`
	BuzzerPort int    `json:"buzzer_port"`
	IsIndex    bool   `json:"is_index"`
}

func (*ClimateConfig) Name() string {
	return ClimateConfigCollectionName
}

func (*ClimateConfig) Schema() *core.Collection {
	collection := core.NewBaseCollection(ClimateConfigCollectionName, ClimateConfigCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = config.device.user.id &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = config.device.user.id`)

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  ConfigCollectionName,
			Name:          "config",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     2,
		},
		&core.NumberField{
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.TextField{
			Name:     "lable",
			Required: true,
		},
		&core.NumberField{
			Name:     "config_id",
			Required: true,
		},
		&core.NumberField{
			Name:     "dht22_port",
			Required: true,
			OnlyInt:  true,
		},
		&core.NumberField{
			Name:     "aqi_port",
			Required: true,
		},
		&core.BoolField{
			Name:     "has_buzzer",
			Required: true,
		},
		&core.NumberField{
			Name:     "buzzer_port",
			Required: true,
		},
		&core.BoolField{
			Name:     "is_index",
			Required: true,
		},
	)

	return collection
}

type LDRConfig struct {
	ID       string `json:"id"`
	SensorId string `json:"sensor_id"`
	Lable    string `json:"lable"`
	Config   string `json:"config"`
	Port     int    `json:"port"`
}

func (*LDRConfig) Name() string {
	return LDRConfigCollectionName
}

func (*LDRConfig) Schema() *core.Collection {
	collection := core.NewBaseCollection(LDRConfigCollectionName, LDRConfigCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = config.device.user.id &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = config.device.user.id`)

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  ConfigCollectionName,
			Name:          "config",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     2,
		},
		&core.NumberField{
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.TextField{
			Name:     "lable",
			Required: true,
		},
		&core.NumberField{
			Name:     "port",
			Required: true,
		},
	)

	return collection
}

type MotionConfig struct {
	ID        string `json:"id"`
	SensorId  string `json:"sensor_id"`
	Lable     string `json:"lable"`
	Config    string `json:"config"`
	Port      int    `json:"port"`
	RelayType int    `json:"relay_type"`
	RelayPort int    `json:"relay_port"`
}

func (*MotionConfig) Name() string {
	return MotionConfigCollectionName
}

func (*MotionConfig) Schema() *core.Collection {
	collection := core.NewBaseCollection(MotionConfigCollectionName, MotionConfigCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = config.device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = config.device.user.id &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = config.device.user.id`)

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  ConfigCollectionName,
			Name:          "config",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     4,
		},
		&core.NumberField{
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.TextField{
			Name:     "lable",
			Required: true,
		},
		&core.NumberField{
			Name:     "port",
			Required: true,
		},
		&core.NumberField{
			Name:     "relay_type",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.NumberField{
			Name:     "relay_port",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
	)

	return collection
}
