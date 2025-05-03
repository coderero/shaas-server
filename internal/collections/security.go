package collections

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	// SecurityCollectionName is the name of the security collection.
	SecurityCollectionName     = "security"
	SecurityLogsCollectionName = "security_logs"
)

type Security struct {
	ID        string `json:"id"`
	Device    string `json:"device"`
	UUID      string `json:"uuid"`
	Timestamp string `json:"timestamp"`
}

func (*Security) Name() string {
	return SecurityCollectionName
}

func (*Security) Schema() *core.Collection {
	collection := core.NewBaseCollection(SecurityCollectionName, SecurityCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
	collection.UpdateRule = types.Pointer(`
        @request.auth.id != '' &&
		@request.auth.id = device.user.id &&
        (@request.body.user:isset = false || @request.body.user = @request.auth.id)
    `)
	collection.DeleteRule = types.Pointer(`@request.auth.id != '' && @request.auth.id = device.user`)

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  DevicesCollectionName,
			Name:          "device",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     1,
		},
		&core.TextField{
			Name:     "uuid",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}

type SecurityLogs struct {
	ID        string `json:"id"`
	Device    string `json:"device"`
	UUID      string `json:"uuid"`
	Level     string `json:"level"`
	Details   string `json:"details"`
	Timestamp string `json:"timestamp"`
}

func (*SecurityLogs) Name() string {
	return SecurityLogsCollectionName
}

func (*SecurityLogs) Schema() *core.Collection {
	collection := core.NewBaseCollection(SecurityLogsCollectionName, SecurityLogsCollectionName)
	collection.ListRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.ViewRule = types.Pointer("@request.auth.id != '' && @request.auth.id = device.user.id")
	collection.CreateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.UpdateRule = types.Pointer("@request.auth.isAdmin = true")
	collection.DeleteRule = types.Pointer("@request.auth.isAdmin = true")

	collection.Fields.Add(
		&core.RelationField{
			CollectionId:  DevicesCollectionName,
			Name:          "device",
			CascadeDelete: true,
			Required:      true,
			MinSelect:     1,
			MaxSelect:     1,
		},
		&core.TextField{
			Name:     "uuid",
			Required: true,
		},
		&core.TextField{
			Name:     "level",
			Required: true,
		},
		&core.TextField{
			Name:     "details",
			Required: true,
		},
		&core.AutodateField{
			Name:     "timestamp",
			OnCreate: true,
		},
	)

	return collection
}
