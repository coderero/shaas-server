package topics

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"coderero.dev/iot/smaas-server/internal/collections"
	"coderero.dev/iot/smaas-server/internal/proto/transporter"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"google.golang.org/protobuf/proto"
)

type Arduino struct {
	app         core.App
	mqttServer  *mqtt.Server
	syncRequest bool
	collections []collections.CollectionDefiner
}

func NewArduino(collections []collections.CollectionDefiner, app core.App, mqttServer *mqtt.Server) *Arduino {
	return &Arduino{
		collections: collections,
		app:         app,
		mqttServer:  mqttServer,
		syncRequest: false,
	}
}
func (a *Arduino) Climate(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	a.app.Logger().Info("climate data received", slog.String("topic", pk.TopicName))
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	var d transporter.ClimateData
	if err := proto.Unmarshal(pk.Payload, &d); err != nil {
		a.app.Logger().Error("failed to unmarshal climate data", slog.String("error", err.Error()))
		return
	}

	record := core.NewRecord(a.getCollection(collections.ClimateCollectionName))
	record.Set("sensor_id", int(d.Id))
	record.Set("device", deviceId)
	record.Set("temperature", d.Temperature)
	record.Set("humidity", d.Humidity)
	record.Set("air_quality", int(d.Aqi))

	a.app.Logger().Info("climate data", slog.String("device_id", deviceId), slog.String("temperature", fmt.Sprintf("%f", d.Temperature)), slog.String("humidity", fmt.Sprintf("%f", d.Humidity)), slog.String("air_quality", fmt.Sprintf("%d", d.Aqi)))

	if err := a.app.Save(record); err != nil {
		return
	}
}

func (a *Arduino) LDR(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	var d transporter.LDRData
	if err := proto.Unmarshal(pk.Payload, &d); err != nil {
		a.app.Logger().Error("failed to unmarshal LDR data", slog.String("error", err.Error()))
		return
	}
	record := core.NewRecord(a.getCollection(collections.LDRCollectionName))
	record.Set("sensor_id", d.Id)
	record.Set("device", deviceId)
	record.Set("ldr_value", d.Value)

	if err := a.app.Save(record); err != nil {
		a.app.Logger().Error("failed to save LDR data", slog.String("error", err.Error()))
		return
	}
}

func (a *Arduino) Relay(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	var d transporter.RelayState
	if err := proto.Unmarshal(pk.Payload, &d); err != nil {
		a.app.Logger().Error("failed to unmarshal relay data", slog.String("error", err.Error()))
		return
	}

	a.syncRequest = true

	var relayId string
	if d.Type == 1 {
		relayId = "relaylowduty001"
	} else {
		relayId = "relayheavyduty1"
	}

	record, err := a.app.FindFirstRecordByFilter(
		a.getCollection(collections.UserPortLablesCollectionName),
		"device = {:device} && port = {:port} && relay = {:relay}",
		dbx.Params{
			"device": deviceId,
			"port":   d.Port,
			"relay":  relayId,
		},
	)

	if err != nil {
		a.app.Logger().Error("failed to find relay record", slog.String("error", err.Error()))
		return
	}

	if record == nil {
		a.app.Logger().Error("failed to find relay record", slog.String("device_id", deviceId), slog.String("port", fmt.Sprintf("%d", d.Port)))
		return
	}

	var state bool
	if d.State == transporter.RelayStateType_ON {
		state = true
	} else {
		state = false
	}

	record.Set("state", state)
	if err := a.app.Save(record); err != nil {
		a.app.Logger().Error("failed to save relay data", slog.String("error", err.Error()))
		return
	}

	a.syncRequest = false
}

func (a *Arduino) FullRelayStateSync(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}
	var d transporter.RelayStateSync
	if err := proto.Unmarshal(pk.Payload, &d); err != nil {
		a.app.Logger().Error("failed to unmarshal relay data", slog.String("error", err.Error()))
		return
	}

	a.syncRequest = true

	records, err := a.app.FindRecordsByFilter(
		a.getCollection(collections.UserPortLablesCollectionName),
		"device = {:device}",
		"",
		10,
		0,
		dbx.Params{
			"device": deviceId,
		},
	)
	if err != nil {
		a.app.Logger().Error("failed to find relay records", slog.String("error", err.Error()))
		return
	}

	if records == nil {
		a.app.Logger().Error("failed to find relay records", slog.String("device_id", deviceId))
		return
	}

	topic := fmt.Sprintf("arduino/%s/relay", deviceId)

	sd := &transporter.RelayState{}

	for _, record := range records {
		relayId := record.GetString("relay")
		port := record.GetInt("port")
		state := record.GetBool("state")

		if relayId == "relaylowduty001" {
			sd.Type = transporter.RelayType_LOW_DUTY
		} else {
			sd.Type = transporter.RelayType_HEAVY_DUTY
		}

		sd.Port = uint32(port)

		if state {
			sd.State = transporter.RelayStateType_ON
		} else {
			sd.State = transporter.RelayStateType_OFF
		}

		payload, err := proto.Marshal(sd)
		if err != nil {
			a.app.Logger().Error("failed to marshal relay data", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
			a.app.Logger().Error("failed to publish relay data", slog.String("error", err.Error()))
			return
		}

		a.app.Logger().Info("published relay data", slog.String("topic", topic), slog.String("device_id", deviceId))
		time.Sleep(100 * time.Millisecond)
	}
	a.app.Logger().Info("full relay state sync", slog.String("topic", pk.TopicName), slog.String("device_id", deviceId))
	a.syncRequest = false
}

func (a *Arduino) Secuirty(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	var d transporter.RfidEnvelope
	if err := proto.Unmarshal(pk.Payload, &d); err != nil {
		a.app.Logger().Error("failed to unmarshal security data", slog.String("error", err.Error()))
		return
	}

	switch d.Payload.(type) {
	case *transporter.RfidEnvelope_RegisterResponse:
		registerResponse := d.GetRegisterResponse()
		// get the security record with id
		securityRecord, err := a.app.FindRecordById(collections.SecurityCollectionName, registerResponse.Id)
		if err != nil {
			a.app.Logger().Error("failed to find security record", slog.String("error", err.Error()))
			return
		}

		if securityRecord == nil {
			a.app.Logger().Error("failed to find security record")
			return
		}
		securityRecord.Set("device", deviceId)
		securityRecord.Set("uuid", hex.EncodeToString(registerResponse.Uid.Value))
		if err := a.app.Save(securityRecord); err != nil {
			a.app.Logger().Error("failed to save security data", slog.String("error", err.Error()))
			return
		}

		a.app.Logger().Info("security data processed", slog.String("topic", pk.TopicName), slog.String("device_id", deviceId))
		return
	default:
		a.app.Logger().Error("failed to unmarshal security data", slog.String("error", "unknown payload type"))
		return
	}
}

func (a *Arduino) RegisterTopics() {
	if err := a.mqttServer.Subscribe("arduino/+/climate", 0, a.Climate); err != nil {
		a.app.Logger().Error("failed to subscribe to climate topic", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Subscribe("arduino/+/ldr", 0, a.LDR); err != nil {
		a.app.Logger().Error("failed to subscribe to LDR topic", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Subscribe("arduino/+/relay", 0, a.Relay); err != nil {
		a.app.Logger().Error("failed to subscribe to relay topic", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Subscribe("arduino/+/relay/full", 0, a.FullRelayStateSync); err != nil {
		a.app.Logger().Error("failed to subscribe to full relay state sync topic", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Subscribe("arduino/+/rfid", 0, a.Secuirty); err != nil {
		a.app.Logger().Error("failed to subscribe to security topic", slog.String("error", err.Error()))
		return
	}

	a.app.OnRecordAfterCreateSuccess(
		collections.SecurityCollectionName,
	).BindFunc(a.securityRegister)
	a.app.OnRecordDeleteExecute(
		collections.SecurityCollectionName,
	).BindFunc(a.securityRevoke)

	a.app.OnRecordAfterCreateSuccess(
		collections.ClimateConfigCollectionName,
		collections.LDRConfigCollectionName,
		collections.MotionConfigCollectionName,
	).BindFunc(a.configHook)

	a.app.OnRecordUpdateExecute(
		collections.UserPortLablesCollectionName,
	).BindFunc(a.relaySwitchHook)

	a.app.OnRecordAfterDeleteSuccess(
		collections.ClimateConfigCollectionName,
		collections.LDRConfigCollectionName,
		collections.MotionConfigCollectionName,
	).BindFunc(a.configResetHook)

	a.app.OnRecordDeleteExecute(
		collections.DevicesCollectionName,
	).BindFunc(a.factoryResetHook)
}

func (a *Arduino) securityRegister(e *core.RecordEvent) error {
	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}

	deviceId := record.GetString("device")
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/rfid", deviceId)

	var d transporter.RfidEnvelope
	d.Payload = &transporter.RfidEnvelope_RegisterRequest{
		RegisterRequest: &transporter.RegisterRequest{
			Id: record.Id,
		},
	}

	payload, err := proto.Marshal(&d)
	if err != nil {
		a.app.Logger().Error("failed to marshal security data", slog.String("error", err.Error()))
		return nil
	}
	if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
		a.app.Logger().Error("failed to publish security data", slog.String("error", err.Error()))
		return nil
	}
	a.app.Logger().Info("published security data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return e.Next()
}

func (a *Arduino) securityRevoke(e *core.RecordEvent) error {
	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}

	deviceId := record.GetString("device")
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	uid := record.GetString("uuid")
	if uid == "" {
		a.app.Logger().Error("failed to get uid from record", slog.String("record_id", record.Id))
		return nil
	}

	uidBytes, err := hex.DecodeString(uid)
	if err != nil {
		a.app.Logger().Error("failed to decode uid", slog.String("error", err.Error()))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/rfid", deviceId)

	var d transporter.RfidEnvelope
	d.Payload = &transporter.RfidEnvelope_RevokeRequest{
		RevokeRequest: &transporter.RevokeRequest{
			Uid: &transporter.UID{
				Value: uidBytes,
			},
		},
	}

	payload, err := proto.Marshal(&d)
	if err != nil {
		a.app.Logger().Error("failed to marshal security data", slog.String("error", err.Error()))
		return nil
	}
	if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
		a.app.Logger().Error("failed to publish security data", slog.String("error", err.Error()))
		return nil
	}
	a.app.Logger().Info("published security data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return e.Next()
}

func (a *Arduino) configHook(e *core.RecordEvent) error {
	a.app.Logger().Info("config hook called")
	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}
	deviceId := record.GetString("device")
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/config", deviceId)

	configPayload := &transporter.ConfigTopic{}

	collectionName := record.Collection().Name
	switch collectionName {
	case collections.ClimateConfigCollectionName:
		configPayload.Payload = &transporter.ConfigTopic_Climate{
			Climate: &transporter.Climate{
				Id:         uint32(record.GetInt("sensor_id")),
				Dht22Port:  uint32(record.GetInt("dht22_port")),
				AqiPort:    uint32(record.GetInt("aqi_port")),
				HasBuzzers: record.GetBool("has_buzzer"),
				BuzzerPort: uint32(record.GetInt("buzzer_port")),
			},
		}
	case collections.LDRConfigCollectionName:
		configPayload.Payload = &transporter.ConfigTopic_Ldr{
			Ldr: &transporter.LDR{
				Id:   uint32(record.GetInt("sensor_id")),
				Port: uint32(record.GetInt("port")),
			},
		}
	case collections.MotionConfigCollectionName:
		var relayType transporter.RelayType
		if record.GetInt("relay_type") == 1 {
			relayType = transporter.RelayType_LOW_DUTY
		} else {
			relayType = transporter.RelayType_HEAVY_DUTY
		}

		configPayload.Payload = &transporter.ConfigTopic_Motion{
			Motion: &transporter.Motion{
				Id:        uint32(record.GetInt("sensor_id")),
				Port:      uint32(record.GetInt("port")),
				RelayPort: uint32(record.GetInt("relay_port")),
				RelayType: relayType,
			},
		}
	default:
		break
	}

	payload, err := proto.Marshal(configPayload)
	if err != nil {
		a.app.Logger().Error("failed to marshal config data", slog.String("error", err.Error()))
		return nil
	}
	if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
		a.app.Logger().Error("failed to publish config data", slog.String("error", err.Error()))
		return nil
	}

	a.app.Logger().Info("published config data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return nil
}

func (a *Arduino) configResetHook(e *core.RecordEvent) error {
	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}
	deviceId := record.GetString("device")
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	sensorID := record.GetInt("sensor_id")
	topic := fmt.Sprintf("arduino/%s/config/remove", deviceId)

	var c *transporter.ConfigRemoval

	switch record.Collection().Name {
	case collections.ClimateConfigCollectionName:
		c = &transporter.ConfigRemoval{
			Payload: &transporter.ConfigRemoval_Climate{
				Climate: &transporter.ClimateRemoval{
					Id: uint32(sensorID),
				},
			},
		}
	case collections.LDRConfigCollectionName:
		c = &transporter.ConfigRemoval{
			Payload: &transporter.ConfigRemoval_Ldr{
				Ldr: &transporter.LDRRemoval{
					Id: uint32(record.GetInt("sensor_id")),
				},
			},
		}

	case collections.MotionConfigCollectionName:
		c = &transporter.ConfigRemoval{
			Payload: &transporter.ConfigRemoval_Motion{
				Motion: &transporter.MotionRemoval{
					Id: uint32(record.GetInt("sensor_id")),
				},
			},
		}
	}

	payload, err := proto.Marshal(c)
	if err != nil {
		a.app.Logger().Error("failed to marshal config data", slog.String("error", err.Error()))
		return nil
	}
	if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
		a.app.Logger().Error("failed to publish config data", slog.String("error", err.Error()))
		return nil
	}

	a.app.Logger().Info("published config data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return e.Next()
}

func (a *Arduino) factoryResetHook(e *core.RecordEvent) error {
	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}
	deviceId := record.Id
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/factory_reset", deviceId)

	// random bytes
	d := []byte{0x00}

	if err := a.mqttServer.Publish(topic, d, false, 0); err != nil {
		a.app.Logger().Error("failed to publish factory reset data", slog.String("error", err.Error()))
		return nil
	}

	a.app.Logger().Info("published factory reset data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return e.Next()
}

func (a *Arduino) relaySwitchHook(e *core.RecordEvent) error {
	if a.syncRequest {
		return e.Next()
	}

	record := e.Record
	if record == nil {
		a.app.Logger().Error("failed to get record from event")
		return nil
	}
	deviceId := record.GetString("device")
	if deviceId == "" {
		a.app.Logger().Error("failed to get device id from record", slog.String("record_id", record.Id))
		return nil
	}

	var t transporter.RelayType
	relayId := record.GetString("relay")
	if relayId == "relaylowduty001" {
		t = transporter.RelayType_LOW_DUTY
	} else {
		t = transporter.RelayType_HEAVY_DUTY
	}

	var s transporter.RelayStateType
	state := record.GetBool("state")
	if state {
		s = transporter.RelayStateType_ON
	} else {
		s = transporter.RelayStateType_OFF
	}

	var d transporter.RelayState
	d.Type = t
	d.State = s
	d.Port = uint32(record.GetInt("port"))
	payload, err := proto.Marshal(&d)
	if err != nil {
		a.app.Logger().Error("failed to marshal relay data", slog.String("error", err.Error()))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/relay", deviceId)
	if err := a.mqttServer.Publish(topic, payload, false, 0); err != nil {
		a.app.Logger().Error("failed to publish relay data", slog.String("error", err.Error()))
		return nil
	}
	a.app.Logger().Info("published relay data", slog.String("topic", topic), slog.String("device_id", deviceId))
	return e.Next()
}

func (*Arduino) getId(topic string) string {
	topicParts := strings.Split(topic, "/")
	if len(topicParts) < 2 {
		return ""
	}

	return topicParts[1]
}

func (a *Arduino) getCollection(collectionName string) *core.Collection {
	for _, c := range a.collections {
		if c.Name() == collectionName {
			return c.Schema()
		}
	}
	return nil
}
