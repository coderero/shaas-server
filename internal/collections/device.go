package collections

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	DevicesCollectionName    = "devices"
	ClimateCollectionName    = "climate"
	LDRCollectionName        = "ldr"
	MotionCollectionName     = "motion"
	RelayCollectionName      = "relay"
	RelayPortsCollectionName = "user_port_lables"
)

type Devices struct {
	ID           string `json:"id"`
	User         string `json:"user"`
	DeviceName   string `json:"device_name"`
	DeviceStatus string `json:"device_status"`
	Timestamp    string `json:"timestamp"`
}

func (*Devices) Name() string {
	return DevicesCollectionName
}

func (*Devices) Schema() *core.Collection {
	collection := core.NewBaseCollection(DevicesCollectionName, DevicesCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user")
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = user")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = user &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = user`)

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  "_pb_users_auth_",
			Name:          "user",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     1,
		},
		&core.TextField{
			Name:     "device_name",
			Required: true,
		},
		&core.TextField{
			Name:     "device_status",
			Required: true,
		},
	)

	return collection
}

type Climate struct {
	ID          string  `json:"id"`
	SensorID    int     `json:"sensor_id"`
	Device      string  `json:"device"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	AirQuality  float64 `json:"air_quality"`
	Timestamp   string  `json:"timestamp"`
}

func (*Climate) Name() string {
	return ClimateCollectionName
}

func (*Climate) Schema() *core.Collection {
	collection := core.NewBaseCollection(ClimateCollectionName, ClimateCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = device.user &&
		(@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = device.user.id`)

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
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.NumberField{
			Name:     "temperature",
			Required: true,
		},
		&core.NumberField{
			Name:     "humidity",
			Required: true,
		},
		&core.NumberField{
			Name:     "air_quality",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}

type LDR struct {
	ID        string  `json:"id"`
	SensorID  int     `json:"sensor_id"`
	Device    string  `json:"device"`
	LDRValue  float64 `json:"ldr_value"`
	Timestamp string  `json:"timestamp"`
}

func (*LDR) Name() string {
	return LDRCollectionName
}

func (*LDR) Schema() *core.Collection {
	collection := core.NewBaseCollection(LDRCollectionName, LDRCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = device.user &&
		(@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = device.user.id`)

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
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.NumberField{
			Name:     "ldr_value",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}

type Motion struct {
	ID          string `json:"id"`
	SensorID    int    `json:"sensor_id"`
	Device      string `json:"device"`
	MotionValue bool   `json:"motion_value"`
	Timestamp   string `json:"timestamp"`
}

func (*Motion) Name() string {
	return MotionCollectionName
}

func (*Motion) Schema() *core.Collection {
	collection := core.NewBaseCollection(MotionCollectionName, MotionCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = device.user &&
		(@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = device.user.id`)

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
			Name:     "sensor_id",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.BoolField{
			Name:     "motion_value",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}

type Relay struct {
	ID       string `json:"id"`
	Type     int    `json:"type"`
	Switches int    `json:"switches"`
}

func (*Relay) Name() string {
	return RelayCollectionName
}

func (*Relay) Schema() *core.Collection {
	collection := core.NewBaseCollection(RelayCollectionName, RelayCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != ''")
	collection.ViewRule = types.Pointer("@request.auth.id != ''")
	collection.CreateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.UpdateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.DeleteRule = types.Pointer("@request.auth.isAdmin = true")

	collection.Fields.Add(
		&core.BoolField{
			Name:     "type",
			Required: true,
		},
		&core.NumberField{
			Name:     "switches",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}

type UserPortLables struct {
	ID        string `json:"id"`
	Device    string `json:"device"`
	Relay     string `json:"relay"`
	State     int    `json:"state"`
	Lable     string `json:"lable"`
	Timestamp string `json:"timestamp"`
}

func (*UserPortLables) Name() string {
	return RelayPortsCollectionName
}

func (*UserPortLables) Schema() *core.Collection {
	collection := core.NewBaseCollection(RelayPortsCollectionName, RelayPortsCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.UpdateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.DeleteRule = types.Pointer("@request.auth.isAdmin = true")

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  RelayCollectionName,
			Name:          "relay",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     1,
		},
		&core.RelationField{
			CollectionId:  "devices",
			Name:          "device",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     6,
		},
		&core.NumberField{
			Name:     "state",
			Required: true,
			OnlyInt:  true,
			System:   true,
		},
		&core.TextField{
			Name:     "lable",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}
