package main

import (
	"context"

	"github.com/al1376kost/pgxlog"

	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

func main() {
	db, err := pgxpool.Connect(context.Background(), "user=postgres dbname=postgres host=postgres sslmode=disable")
	if err != nil {
		log.Fatal("Can't connect to postgresql database:", err)
	}
	defer db.Close()
	hook := pgxlog.NewAsyncHook(db, map[string]interface{}{"this": "is logged every time"})
	defer hook.Flush()
	log.AddHook(hook)
	log.Info("some logging message")
}
