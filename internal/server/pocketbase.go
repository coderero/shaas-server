package server

import (
	"coderero.dev/iot/smaas-server/internal/collections"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pocketbase/pocketbase"
)

type PocketBase struct {
	app         *pocketbase.PocketBase
	collections []collections.CollectionDefiner
}

func NewPocketBase() *PocketBase {
	return &PocketBase{
		app: pocketbase.New(),
		collections: []collections.CollectionDefiner{
			&collections.Devices{},
			&collections.WifiCredentials{},
			&collections.Climate{},
			&collections.LDR{},
			&collections.Relay{},
			&collections.Security{},
			&collections.ClimateConfig{},
			&collections.LDRConfig{},
			&collections.Relay{},
			&collections.UserPortLables{},
			&collections.MotionConfig{},
		},
	}
}
