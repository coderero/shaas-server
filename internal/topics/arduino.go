package topics

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"coderero.dev/iot/smaas-server/internal/collections"
	"coderero.dev/iot/smaas-server/internal/proto/transporter"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/pocketbase/pocketbase/core"
	"google.golang.org/protobuf/proto"
)

type Arduino struct {
	app         core.App
	logger      *slog.Logger
	rootTopic   string
	mqttServer  *mqtt.Server
	collections []collections.CollectionDefiner
}

func NewArduino(rootTopic string, collections []collections.CollectionDefiner, app core.App, mqttServer *mqtt.Server) *Arduino {
	return &Arduino{
		collections: collections,
		app:         app,
		logger:      app.Logger(),
		rootTopic:   rootTopic,
		mqttServer:  mqttServer,
	}
}

func (a *Arduino) Sensors(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	sensorData := &transporter.SensorData{}
	if err := proto.Unmarshal(pk.Payload, sensorData); err != nil {
		a.logger.Error("failed to unmarshal sensor data", slog.String("error", err.Error()))
		return
	}

	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.logger.Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	switch sensorData.Data.(type) {
	case *transporter.SensorData_Climate:

		climateCollection := a.getCollection(collections.ClimateCollectionName)
		if climateCollection == nil {
			a.logger.Error("failed to get climate config collection")
			return
		}

		climateRecord := core.NewRecord(climateCollection)
		climateRecord.Set("sensor_id", sensorData.GetClimate().Id)
		climateRecord.Set("device", deviceId)
		climateRecord.Set("temperature", sensorData.GetClimate().Temperature)
		climateRecord.Set("humidity", sensorData.GetClimate().Humidity)
		climateRecord.Set("air_quality", sensorData.GetClimate().Aqi)

		if err := a.app.Save(climateRecord); err != nil {
			a.logger.Error("failed to save climate data", slog.String("error", err.Error()))
			return
		}
		a.logger.Info("saved climate data",
			slog.String("device_id", deviceId),
			slog.Float64("temperature", float64(sensorData.GetClimate().Temperature)),
			slog.Float64("humidity", float64(sensorData.GetClimate().Humidity)),
			slog.Float64("air_quality", float64(sensorData.GetClimate().Aqi)),
		)

	case *transporter.SensorData_Ldr:
		ldrCollection := a.getCollection(collections.LDRCollectionName)
		if ldrCollection == nil {
			a.logger.Error("failed to get LDR config collection")
			return
		}
		ldrRecord := core.NewRecord(ldrCollection)
		ldrRecord.Set("sensor_id", sensorData.GetLdr().Id)
		ldrRecord.Set("device", deviceId)
		ldrRecord.Set("ldr_value", sensorData.GetLdr().LdrValue)
		if err := a.app.Save(ldrRecord); err != nil {
			a.logger.Error("failed to save LDR data", slog.String("error", err.Error()))
			return
		}

		a.logger.Info("saved LDR data", slog.String("device_id", deviceId), slog.Float64("ldr_value", float64(sensorData.GetLdr().LdrValue)))
	case *transporter.SensorData_Motion:
		motionCollection := a.getCollection(collections.MotionCollectionName)
		if motionCollection == nil {
			a.logger.Error("failed to get motion config collection")
			return
		}

		motionRecord := core.NewRecord(motionCollection)
		motionRecord.Set("sensor_id", sensorData.GetMotion().Id)
		motionRecord.Set("device", deviceId)
		motionRecord.Set("motion_value", sensorData.GetMotion().MotionDetected)
		if err := a.app.Save(motionRecord); err != nil {
			a.logger.Error("failed to save motion data", slog.String("error", err.Error()))
			return
		}

		a.logger.Info("saved motion data", slog.String("device_id", deviceId), slog.Bool("motion_value", sensorData.GetMotion().MotionDetected))

	}

	a.logger.Info("sensor data received", slog.String("topic", pk.TopicName), slog.String("device_id", deviceId))
}

func (a *Arduino) Auth(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.logger.Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}

	securityCollection := a.getCollection(collections.SecurityCollectionName)
	if securityCollection == nil {
		a.logger.Error("failed to get security config collection")
		return
	}

	pbData := &transporter.SecurityCommand{}
	pbStatus := &transporter.SecurityStatus{
		Status: true,
	}

	if err := proto.Unmarshal(pk.Payload, pbData); err != nil {
		a.logger.Error("failed to unmarshal security data", slog.String("error", err.Error()))
		pbStatus.Status = false

		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}

	securityRecord, err := a.app.FindFirstRecordByFilter(
		collections.SecurityCollectionName,
		fmt.Sprintf("device_id = %s", pbData.DeviceId),
	)
	if err != nil {
		a.logger.Error("failed to find security record", slog.String("error", err.Error()))
		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}

	if securityRecord == nil {
		a.logger.Error("failed to find security record")
		pbStatus.Status = false
		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}

	// decode the security uuid from hex to string
	pbByte, err := hex.DecodeString(securityRecord.GetString("uuid"))
	if err != nil {
		a.logger.Error("failed to decode security uuid", slog.String("error", err.Error()))
		pbStatus.Status = false
		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}

	if bytes.Equal(pbData.Uid, pbByte) {
		pbStatus.Status = true
	} else {
		pbStatus.Status = false
	}
	pbStatusBytes, err := proto.Marshal(pbStatus)
	if err != nil {
		a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Publish(
		fmt.Sprintf("arduino/%s/auth", deviceId),
		pbStatusBytes,
		false,
		0,
	); err != nil {
		a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
		return
	}

	a.logger.Info("security data received", slog.String("topic", pk.TopicName), slog.String("device_id", deviceId))
}

func (a *Arduino) RegisterUid(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
	deviceId := a.getId(pk.TopicName)
	if deviceId == "" {
		a.logger.Error("failed to get device id from topic", slog.String("topic", pk.TopicName))
		return
	}
	securityCollection := a.getCollection(collections.SecurityCollectionName)
	if securityCollection == nil {
		a.logger.Error("failed to get security config collection")
		return
	}

	pbData := &transporter.SecurityCommand{}
	pbStatus := &transporter.SecurityStatus{
		Status: true,
	}

	if err := proto.Unmarshal(pk.Payload, pbData); err != nil {
		a.logger.Error("failed to unmarshal security data", slog.String("error", err.Error()))
		pbStatus.Status = false

		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}

	securityRecord, err := a.app.FindRecordById(collections.SecurityCollectionName, pbData.GetRequestId())
	if err != nil {
		a.logger.Error("failed to find security record", slog.String("error", err.Error()))
		pbStatus.Status = false
		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}
	securityRecord.Set("device", deviceId)
	securityRecord.Set("uuid", hex.EncodeToString(pbData.Uid))

	if err := a.app.Save(securityRecord); err != nil {
		a.logger.Error("failed to save security data", slog.String("error", err.Error()))
		pbStatus.Status = false
		pbStatusBytes, err := proto.Marshal(pbStatus)
		if err != nil {
			a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
			return
		}
		if err := a.mqttServer.Publish(
			fmt.Sprintf("arduino/%s/auth", deviceId),
			pbStatusBytes,
			false,
			0,
		); err != nil {
			a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
			return
		}
		return
	}
	pbStatus.Status = true
	pbStatusBytes, err := proto.Marshal(pbStatus)
	if err != nil {
		a.logger.Error("failed to marshal security status", slog.String("error", err.Error()))
		return
	}
	if err := a.mqttServer.Publish(
		fmt.Sprintf("arduino/%s/auth", deviceId),
		pbStatusBytes,
		false,
		0,
	); err != nil {
		a.logger.Error("failed to publish security status", slog.String("error", err.Error()))
		return
	}
	a.logger.Info("security data received", slog.String("topic", pk.TopicName), slog.String("device_id", deviceId))
}

func (a *Arduino) RegisterTopics() {
	a.mqttServer.Subscribe(fmt.Sprintf("%s/+/sensors", a.rootTopic), 0, a.Sensors)
	a.mqttServer.Subscribe(fmt.Sprintf("%s/+/auth", a.rootTopic), 0, a.Auth)
	a.mqttServer.Subscribe(fmt.Sprintf("%s/+/register", a.rootTopic), 0, a.RegisterUid)

	a.app.OnRecordAfterCreateSuccess(
		collections.SecurityCollectionName,
	).BindFunc(a.securityRegister)
	a.app.OnRecordAfterUpdateSuccess(
		collections.SecurityCollectionName,
	).BindFunc(a.securityRegister)

	a.app.OnRecordAfterCreateSuccess(
		collections.ConfigCollectionName,
		collections.ClimateConfigCollectionName,
		collections.LDRConfigCollectionName,
		collections.MotionConfigCollectionName,
		collections.RelayConfigCollectionName,
	).BindFunc(a.configHook)

	a.app.OnRecordAfterUpdateSuccess(
		collections.ConfigCollectionName,
		collections.ClimateConfigCollectionName,
		collections.LDRConfigCollectionName,
		collections.MotionConfigCollectionName,
		collections.RelayConfigCollectionName,
	).BindFunc(a.configHook)
	a.app.OnRecordDelete(
		collections.DevicesCollectionName,
		collections.ConfigCollectionName,
	).BindFunc(a.configResetHook)
}

func (a *Arduino) securityRegister(e *core.RecordEvent) error {
	securityCollection := a.getCollection(collections.SecurityCollectionName)
	if securityCollection == nil {
		a.logger.Error("failed to get security config collection")
		return nil
	}

	pbData := &transporter.RegistrationCommand{
		DeviceId:  e.Record.GetString("device"),
		RequestId: e.Record.Id,
	}

	pbDataBytes, err := proto.Marshal(pbData)
	if err != nil {
		a.logger.Error("failed to marshal security data", slog.String("error", err.Error()))
		return nil
	}
	topic := fmt.Sprintf("arduino/%s/register", e.Record.GetString("device"))
	if err := a.mqttServer.Publish(topic, pbDataBytes, false, 0); err != nil {
		a.logger.Error("failed to publish security data", slog.String("error", err.Error()))
		return nil
	}
	a.logger.Info("published security data", slog.String("topic", topic))
	return nil
}

func (a *Arduino) configHook(e *core.RecordEvent) error {
	configData := &transporter.ConfigData{}
	configCollection := a.getCollection(collections.ConfigCollectionName)
	if configCollection == nil {
		a.logger.Error("failed to get config collection")
		return nil
	}

	climateConfigCollection := a.getCollection(collections.ClimateConfigCollectionName)
	if climateConfigCollection == nil {
		a.logger.Error("failed to get climate config collection")
		return nil
	}

	ldrConfigCollection := a.getCollection(collections.LDRConfigCollectionName)
	if ldrConfigCollection == nil {
		a.logger.Error("failed to get LDR config collection")
		return nil
	}

	motionConfigCollection := a.getCollection(collections.MotionConfigCollectionName)
	if motionConfigCollection == nil {
		a.logger.Error("failed to get motion config collection")
		return nil
	}

	relayConfigCollection := a.getCollection(collections.RelayConfigCollectionName)
	if relayConfigCollection == nil {
		a.logger.Error("failed to get relay config collection")
		return nil
	}

	var id string
	var deviceId string

	collectionName := e.Record.Collection().Name
	if collectionName == collections.ConfigCollectionName {
		id = e.Record.Id
		deviceId = e.Record.GetString("device")
	} else {
		id = e.Record.GetString("config")
		config, err := a.app.FindRecordById(configCollection, id)
		if err != nil {
			a.logger.Error("failed to find device record", slog.String("error", err.Error()))
			return nil
		}
		deviceId = config.Id
	}
	if id == "" {
		a.logger.Error("failed to get config id from record", slog.String("record_id", e.Record.Id))
		return nil
	}

	if deviceId == "" {
		a.logger.Error("failed to find device record", slog.String("device_id", deviceId))
		return nil
	}

	configData.Version = uint32(e.Record.GetInt("version"))

	climateRecords, err := a.app.FindRecordsByFilter(
		climateConfigCollection,
		"config = "+id,
		"-sensor_id",
		5,
		0,
	)

	if err != nil {
		a.logger.Error("failed to get climate config records", slog.String("error", err.Error()))
		return nil
	}
	ldrRecords, err := a.app.FindRecordsByFilter(
		ldrConfigCollection,
		"config = "+id,
		"-sensor_id",
		5,
		0,
	)
	if err != nil {
		a.logger.Error("failed to get LDR config records", slog.String("error", err.Error()))
		return nil
	}

	motionRecords, err := a.app.FindRecordsByFilter(
		motionConfigCollection,
		"config = "+id,
		"-sensor_id",
		5,
		0,
	)

	if err != nil {
		a.logger.Error("failed to get motion config records", slog.String("error", err.Error()))
		return nil
	}

	relayRecords, err := a.app.FindRecordsByFilter(
		relayConfigCollection,
		"config = "+id,
		"-sensor_id",
		5,
		0,
	)

	if err != nil {
		a.logger.Error("failed to get relay config records", slog.String("error", err.Error()))
		return nil
	}

	climates := []*transporter.Climate{}
	ldrs := []*transporter.Ldr{}
	motions := []*transporter.Motion{}
	relays := []*transporter.Relay{}

	for _, record := range climateRecords {
		climate := &transporter.Climate{
			Id:         uint32(record.GetInt("sensor_id")),
			Dht22Port:  uint32(record.GetInt("dht22_port")),
			AqiPort:    uint32(record.GetInt("aqi_port")),
			HasBuzzer:  record.GetBool("has_buzzer"),
			BuzzerPort: uint32(record.GetInt("buzzer_port")),
		}
		climates = append(climates, climate)
	}

	for _, record := range ldrRecords {
		ldr := &transporter.Ldr{
			Id:   uint32(record.GetInt("sensor_id")),
			Port: uint32(record.GetInt("port")),
		}
		ldrs = append(ldrs, ldr)
	}

	for _, record := range motionRecords {
		motion := &transporter.Motion{
			Id:      uint32(record.GetInt("sensor_id")),
			Port:    uint32(record.GetInt("port")),
			RelayId: uint32(record.GetInt("relay_id")),
		}
		motions = append(motions, motion)
	}

	for _, record := range relayRecords {
		relay := &transporter.Relay{
			Id:   uint32(record.GetInt("sensor_id")),
			Type: transporter.RelayType(record.GetInt("type")),
		}
		relays = append(relays, relay)
	}

	configData.ClimateSize = uint32(len(climates))
	configData.Climates = climates
	configData.LdrSize = uint32(len(ldrs))
	configData.Ldrs = ldrs
	configData.MotionSize = uint32(len(motions))
	configData.Motions = motions
	configData.RelaySize = uint32(len(relays))
	configData.Relays = relays

	configDataBytes, err := proto.Marshal(configData)
	if err != nil {
		a.logger.Error("failed to marshal config data", slog.String("error", err.Error()))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/config", deviceId)
	if err := a.mqttServer.Publish(topic, configDataBytes, false, 0); err != nil {
		a.logger.Error("failed to publish config data", slog.String("error", err.Error()))
		return nil
	}

	a.logger.Info("published config data")
	return nil
}

func (a *Arduino) configResetHook(e *core.RecordEvent) error {
	configData := &transporter.ConfigData{
		Version:     0,
		ClimateSize: 0,
		Climates:    nil,
		LdrSize:     0,
		Ldrs:        nil,
		MotionSize:  0,
		Motions:     nil,
		RelaySize:   0,
		Relays:      nil,
	}

	var id string
	var deviceId string

	colletionName := e.Record.Collection().Name
	if colletionName == collections.ConfigCollectionName {
		id = e.Record.GetString("id")
		deviceId = e.Record.GetString("device_id")
	} else {
		id = e.Record.GetString("config")
		config, err := a.app.FindRecordById(collections.ConfigCollectionName, id)
		if err != nil {
			a.logger.Error("failed to find device record", slog.String("error", err.Error()))
			return nil
		}
		deviceId = config.Id
	}
	if id == "" {
		a.logger.Error("failed to get config id from record", slog.String("record_id", e.Record.Id))
		return nil
	}
	if deviceId == "" {
		a.logger.Error("failed to find device record", slog.String("device_id", deviceId))
		return nil
	}
	configDataBytes, err := proto.Marshal(configData)
	if err != nil {
		a.logger.Error("failed to marshal config data", slog.String("error", err.Error()))
		return nil
	}

	topic := fmt.Sprintf("arduino/%s/config", deviceId)
	if err := a.mqttServer.Publish(topic, configDataBytes, false, 0); err != nil {
		a.logger.Error("failed to publish config data", slog.String("error", err.Error()))
		return nil
	}
	a.logger.Info("published config data")
	return nil
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
