# PostgreSQL (github.com/jackc/pgx/pgxpool) Hook for [Logrus](https://github.com/sirupsen/logrus)

Use this hook to send your logs to [postgresql](http://postgresql.org) server.

## Usage

The hook must be configured with:

* A postgresql db connection (*`*pgxpool`)
* an optional hash with extra global fields. These fields will be included in all messages sent to postgresql
* recommend using [TimescaleDB](https://www.timescale.com/)

```go
package main

import (
	"context"

	"github.com/al1376kost/pgxlog"

	"github.com/jackc/pgx/v4/pgxpool"
	log "github.com/sirupsen/logrus"
)

func main() {
	db, err := pgxpool.Connect(context.Background(), "dbname=DBNAMEhost=HOST_IP port=HOST_PORT user=postgres password=postgres")
	if err != nil {
		log.Fatal("Can't connect to postgresql database:", err)
	}
	defer db.Close()
	hook := pgxlog.NewHook(db, map[string]interface{}{"this": "is logged every time"})
	defer hook.Flush()
	log.AddHook(hook)
	log.Info("some logging message")
}

```

### Asynchronous logger (recomended) in examples




