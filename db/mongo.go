package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func InitMongoDB() *mongo.Client {
	logger := logs.GetLogger()

	dbURL := fmt.Sprintf("mongodb+srv://%s:%s@%s/",
		config.GetString("mongodb.user"),
		url.PathEscape(config.GetString("mongodb.password")),
		config.GetString("mongodb.host"))
	client, err := mongo.Connect(options.Client().ApplyURI(dbURL))
	if err != nil {
		logger.Fatal("Error connecting to mongo client", zap.Error(err))
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	logger.Info("Connected to MongoDB")

	return client
}
