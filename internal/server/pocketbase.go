package server

import (
	"os"

	"coderero.dev/iot/smaas-server/internal/collections"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
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
			&collections.Climate{},
			&collections.Motion{},
			&collections.LDR{},
			&collections.Relay{},
			&collections.Security{},
			&collections.Config{},
			&collections.ClimateConfig{},
			&collections.LDRConfig{},
			&collections.Relay{},
			&collections.UserPortLables{},
			&collections.MotionConfig{},
		},
	}
}

func (pb *PocketBase) RegisterMigrations() {
	migratecmd.MustRegister(pb.app, pb.app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	pb.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		err := pb.collectionMigration()

		se.Next()
		return err
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

func (pb *PocketBase) collectionMigration() error {
	err := pb.setupSuperuser(pb.app)
	if err != nil {
		if err.Error() != "record already exists" {
			return nil
		}
		return err
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

func (pb *PocketBase) GetCollectionsNames() []collections.CollectionDefiner {
	return pb.collections
}
