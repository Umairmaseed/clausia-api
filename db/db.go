package db

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudflare/cfssl/log"
	"github.com/goledgerdev/goprocess-api/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	database *DB
	once     sync.Once
)

type DB struct {
	Client *mongo.Client
}

// GetDB returns a singleton instance of GMongo client to perform actions in Mongo
func GetDB() *DB {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		user := os.Getenv(env.MONGO_USER)
		passwd := os.Getenv(env.MONGO_PWD)
		mongourl := os.Getenv(env.MONGO_URL)

		if user == "" || passwd == "" || mongourl == "" {
			log.Error(fmt.Sprintf("Missing env vars %s, %s or %s.", env.MONGO_USER, env.MONGO_PWD, env.MONGO_URL))
			return
		}

		var uri string = fmt.Sprintf("mongodb://%s", mongourl)

		credential := options.Credential{
			AuthSource: "admin",
			Username:   user,
			Password:   passwd,
		}

		clientOptions := options.Client().ApplyURI(uri).SetAuth(credential)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Error("failed to connect to db:", err.Error())
			return
		}

		database = &DB{
			Client: client,
		}
	})
	return database
}

func (db *DB) StartSession() (mongo.Session, error) {
	return db.Database().Client().StartSession()
}

func (db *DB) Database() *mongo.Database {
	return db.Client.Database(os.Getenv(env.DATABASE_NAME))
}
