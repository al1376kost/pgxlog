package main

import (
	"context"

	"github.com/al1376kost/pgxlog"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

// Config config params
type Config struct {
	DatabaseURL string `toml:"database_url"`
}

func main() {
	config := &Config{}
	if _, err := toml.DecodeFile("./config/config.toml", config); err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.Connect(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal("Can't connect to postgresql database:", err)
	}
	defer db.Close()
	hook := pgxlog.NewAsyncHook(db, map[string]interface{}{"this": "is logged every time"})
	defer hook.Flush()
	log.AddHook(hook)
	log.Info("some logging message")
}
