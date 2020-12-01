package pgxlog

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

// Config ...
type Config struct {
	DatabaseURL string `toml:"database_url"`
}

func TestHooks(t *testing.T) {

	var config Config
	if _, err := toml.DecodeFile("config/config.toml", config); err != nil {
		log.Fatal(err)
	}
	pool, err := pgxpool.Connect(context.Background(), config.DatabaseURL)
	if err != nil {
		t.Fatal("Can't connect to postgresql test database:", err)
	}
	defer pool.Close()

	hooks := map[string]interface {
		logrus.Hook
		Blacklist([]string)
		AddFilter(filter)
	}{
		"Hook":      NewHook(pool, map[string]interface{}{}),
		"AsyncHook": NewAsyncHook(pool, map[string]interface{}{}),
	}

	for name, hook := range hooks {
		t.Run(name, func(t *testing.T) {
			hook.Blacklist([]string{"filterMe"})
			hook.AddFilter(func(entry *logrus.Entry) *logrus.Entry {
				if _, ok := entry.Data["ignore"]; ok {
					// ignore entry
					entry = nil
				}
				return entry
			})

			log := logrus.New()
			log.Out = ioutil.Discard
			log.Level = logrus.DebugLevel
			log.Hooks.Add(hook)

			if h, ok := hook.(*AsyncHook); ok {
				h.FlushEvery(100 * time.Millisecond)
			}

			// Purge our test DB
			_, err = pool.Exec(context.Background(), "DELETE FROM adm.logs_test")
			if err != nil {
				t.Fatal("Can't purge DB:", err)
			}

			msg := "test message\nsecond line"
			errMsg := "some error occurred"

			var wg sync.WaitGroup

			messages := []*logrus.Entry{
				{
					Logger:  log,
					Data:    logrus.Fields{"withField": "1", "user": "123"},
					Level:   logrus.ErrorLevel,
					Caller:  &runtime.Frame{Function: "somefunc"},
					Message: errMsg,
				},
				{
					Logger:  log,
					Data:    logrus.Fields{"withField": "2", "filterMe": "1"},
					Level:   logrus.InfoLevel,
					Caller:  &runtime.Frame{Function: "somefunc"},
					Message: msg,
				},
				{
					Logger:  log,
					Data:    logrus.Fields{"withField": "3"},
					Level:   logrus.DebugLevel,
					Caller:  &runtime.Frame{Function: "somefunc"},
					Message: msg,
				},
				{
					Logger:  log,
					Data:    logrus.Fields{"ignore": "me"},
					Level:   logrus.InfoLevel,
					Caller:  &runtime.Frame{Function: "somefunc"},
					Message: msg,
				},
			}

			for _, entry := range messages {
				wg.Add(1)
				go func(e *logrus.Entry) {
					defer wg.Done()
					switch e.Level {
					case logrus.DebugLevel:
						e.Debug(e.Message)
					case logrus.InfoLevel:
						e.Info(e.Message)
					case logrus.ErrorLevel:
						e.Error(e.Message)
					default:
						t.Error("unknown level:", e.Level)
					}
				}(entry)
			}
			wg.Wait()

			if h, ok := hook.(*AsyncHook); ok {
				h.Flush()
			}

			// Check results in DB
			var (
				data        *json.RawMessage
				addDateTime time.Time
				level       logrus.Level
				message     string
			)
			rows, err := pool.Query(context.Background(), "SELECT add_date_time, level_id, message, message_data FROM adm.logs_test LIMIT 100")
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var numRows int
			for rows.Next() {
				numRows++
				err := rows.Scan(&addDateTime, &level, &message, &data)
				if err != nil {
					t.Fatal(err)
				}

				var expectedData map[string]interface{}
				var expectedMsg string

				switch level {
				case logrus.ErrorLevel:
					expectedMsg = errMsg
					expectedData = map[string]interface{}{
						"withField": "1",
						"user":      "123",
					}

				case logrus.InfoLevel:
					expectedMsg = msg
					expectedData = map[string]interface{}{
						"withField": "2",
						// "filterme" should be filtered
					}
				case logrus.DebugLevel:
					expectedMsg = msg
					expectedData = map[string]interface{}{
						"withField": "3",
					}
				default:
					t.Error("Unknown log level:", level)
				}

				if message != expectedMsg {
					t.Errorf("Expected message to be %q, got %q\n", expectedMsg, message)
				}

				var storedData map[string]interface{}
				if err := json.Unmarshal(*data, &storedData); err != nil {
					t.Fatal("Can't unmarshal data from DB: ", err)
				}
				if !reflect.DeepEqual(expectedData, storedData) {
					t.Errorf("Expected stored data to be %v, got %v\n", expectedData, storedData)
				}

			}
			err = rows.Err()
			if err != nil {
				t.Fatal(err)
			}
			if len(messages)-1 != numRows {
				t.Errorf("Expected %d rows, got %d\n", len(messages), numRows)
			}

		})
	}
}
