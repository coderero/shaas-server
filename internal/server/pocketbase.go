package server

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
)

type PocketBase struct {
	app *pocketbase.PocketBase
}

func New() *PocketBase {
	return &PocketBase{
		app: pocketbase.New(),
	}
}

func (pb *PocketBase) RegisterMigrations() {
	migrations.Register(func(app core.App) error {
		err := pb.setupSuperuser(app)
		if err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}

func (pb *PocketBase) RegisterRoutes() {
	pb.app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})
}

func (pb *PocketBase) setupSuperuser(app core.App) error {
	superuser, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	record := core.NewRecord(superuser)
	record.Set("email", os.Getenv("SUPERUSER_EMAIL"))
	record.Set("password", os.Getenv("SUPERUSER_PASSWORD"))

	return app.Save(record)
}
