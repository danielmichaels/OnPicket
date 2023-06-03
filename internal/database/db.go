package database

import (
	"context"
	"fmt"
	"github.com/danielmichaels/onpicket/internal/config"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

const (
	ScanCollection = "scans"
)

func InitDatabase(db string) (*mongo.Database, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.AppConfig()
	creds := options.Credential{
		Username:   cfg.Db.DbUsername,
		Password:   cfg.Db.DbPassword,
		AuthSource: cfg.Db.DbNameAuth,
	}

	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d",
		cfg.Db.DbUsername,
		cfg.Db.DbPassword,
		cfg.Db.DbConnectionString,
		cfg.Db.DbPort,
	)
	opts := options.Client().ApplyURI(uri).SetAuth(creds)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to mongo instance")
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error().Err(err).Msg("failed to ping mongo instance")
		return nil, err
	}

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = client.Database(db).
		Collection(ScanCollection).
		Indexes().
		CreateMany(context.TODO(), indexes)
	if err != nil {
		log.Error().Err(err).Interface("indexes", indexes).Msg("failed to create indexes")
		return nil, err
	}
	log.Info().Msg("successfully connected and pinged the database.")
	return client.Database(db), nil
}
