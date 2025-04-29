package server

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"coderero.dev/iot/smaas-server/internal/collections"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func (pb *PocketBase) RegisterMigrations() {
	migratecmd.MustRegister(pb.app, pb.app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	pb.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		err := pb.collectionMigration()

		se.Next()
		return err
	})

	pb.app.OnRecordAfterCreateSuccess("devices").BindFunc(func(e *core.RecordEvent) error {
		device := e.Record
		for i := 1; i <= 4; i++ {
			relayPort := pb.newRelayPort(i, 1, device.Id)
			err := e.App.Save(relayPort)
			if err != nil {
				return err
			}
		}

		for i := 1; i <= 2; i++ {
			relayPort := pb.newRelayPort(i, 2, device.Id)
			err := e.App.Save(relayPort)
			if err != nil {
				return err
			}
		}
		return e.Next()
	})

	pb.app.OnRecordCreateExecute("climate_config", "ldr_config", "motion_config").BindFunc(func(e *core.RecordEvent) error {
		record := e.Record
		sensorId := record.GetInt("sensor_id")
		duplicateRecord, err := e.App.FindFirstRecordByFilter(
			record.Collection().Name,
			"sensor_id = {:sensor} && device = {:device}",
			dbx.Params{
				"sensor": sensorId,
				"device": record.GetString("device"),
			},
		)

		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		if duplicateRecord != nil {
			return fmt.Errorf("record with sensor_id %d already exists", sensorId)
		}

		return e.Next()
	})

	pb.app.OnRecordCreateExecute("climate_config", "ldr_config", "motion_config").BindFunc(func(e *core.RecordEvent) error {
		record := e.Record

		switch record.Collection().Name {
		case "climate_config":
			dht_port := record.GetInt("dht22_port")
			aqi_port := record.GetInt("aqi_port")

			duplicateRecord, err := e.App.FindFirstRecordByFilter(
				record.Collection().Name,
				"(dht22_port = {:dht} || aqi_port = {:aqi}) && device = {:device}",
				dbx.Params{
					"dht":    dht_port,
					"aqi":    aqi_port,
					"device": record.GetString("device"),
				},
			)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}

			if duplicateRecord != nil {
				return fmt.Errorf("record with dht_port %d or aqi_port %d already exists", dht_port, aqi_port)
			}
		default:
			portId := record.GetInt("port")
			duplicateRecord, err := e.App.FindFirstRecordByFilter(
				record.Collection().Name,
				"port = {:port} && device = {:device}",
				dbx.Params{
					"port":   portId,
					"device": record.GetString("device"),
				},
			)

			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}
			if duplicateRecord != nil {
				return fmt.Errorf("record with port_id %d already exists", portId)
			}
		}
		return e.Next()
	})
}

func (pb *PocketBase) Start() error {
	if err := pb.app.Start(); err != nil {
		return err
	}
	return nil
}

func (pb *PocketBase) RegisterRoutes() {
	pb.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})
}

func (pb *PocketBase) GetCollectionsNames() []collections.CollectionDefiner {
	return pb.collections
}

func (pb *PocketBase) collectionMigration() error {
	err := pb.setupSuperuser(pb.app)
	if err != nil {
		if err.Error() == "record already exists" {
			return err
		}
	}

	for _, collection := range pb.collections {
		_, err := pb.app.FindCollectionByNameOrId(collection.Name())
		if err != nil {
			collectionDefiner := collection.Schema()
			if err := pb.app.Save(collectionDefiner); err != nil {
				return err
			}
		}
	}

	err = pb.setupRelays(pb.app)
	if err != nil {
		if err.Error() == "id: Value must be unique" {
			return err
		}
	}

	return nil
}

func (pb *PocketBase) setupSuperuser(app core.App) error {
	superuser, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	record := core.NewRecord(superuser)
	record.Set("email", os.Getenv("ADMIN_EMAIL"))
	record.Set("password", os.Getenv("ADMIN_PASSWORD"))

	return app.Save(record)
}

func (pb *PocketBase) setupRelays(app core.App) error {
	relay, err := app.FindCollectionByNameOrId(collections.RelayCollectionName)
	if err != nil {
		return err
	}
	relayHeavyDuty := core.NewRecord(relay)
	relayHeavyDuty.Set("id", "relayheavyduty1")
	relayHeavyDuty.Set("type", 2)
	relayHeavyDuty.Set("switches", 2)

	relayLowDuty := core.NewRecord(relay)
	relayLowDuty.Set("id", "relaylowduty001")
	relayLowDuty.Set("type", 1)
	relayLowDuty.Set("switches", 4)

	err = pb.app.Save(relayHeavyDuty)
	if err != nil {

		return err
	}

	err = pb.app.Save(relayLowDuty)
	if err != nil {
		return err
	}

	return nil
}

func (pb *PocketBase) newRelayPort(port int, typeId int, deviceId string) *core.Record {
	userPortLablesCollection, err := pb.app.FindCollectionByNameOrId(collections.UserPortLablesCollectionName)
	if err != nil {
		return nil
	}

	var relayId string
	if typeId == 1 {
		relayId = "relaylowduty001"
	} else {
		relayId = "relayheavyduty1"
	}
	record := core.NewRecord(userPortLablesCollection)
	record.Set("device", deviceId)
	record.Set("relay", relayId)
	record.Set("port", port)
	record.Set("lable", fmt.Sprintf("Relay %d", port))
	record.Set("state", false)

	return record
}
