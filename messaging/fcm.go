package messaging

import (
	"context"
	"encoding/json"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/skb1129/go-utils/config"
	"github.com/skb1129/go-utils/logs"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type FCMPayload struct {
	Title  string
	Body   string
	Data   map[string]string
	Silent bool
}

type FCM struct {
	client *messaging.Client
}

func NewFCM() *FCM {
	logger := logs.GetLogger()
	serviceAccount := config.GetMap("gcp.credentials")
	credentialsJSON, err := json.Marshal(serviceAccount)
	if err != nil {
		logger.Fatal("Failed to marshal service account credentials", zap.Error(err))
	}
	options := option.WithAuthCredentialsJSON(option.ServiceAccount, credentialsJSON)
	app, err := firebase.NewApp(context.TODO(), nil, options)
	if err != nil {
		logger.Fatal("Failed to initialize Firebase App", zap.Error(err))
	}
	client, err := app.Messaging(context.TODO())
	if err != nil {
		logger.Fatal("Failed to initialize Firebase Messaging client", zap.Error(err))
	}
	return &FCM{client: client}
}

func (c *FCM) SendNotifications(ctx context.Context, tokens []string, payload *FCMPayload) map[string]int {
	failureCount := 0
	successCount := 0
	for i := 0; i < len(tokens); i += 500 {
		end := i + 500
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]
		msg := &messaging.MulticastMessage{
			Tokens: batch,
			Data:   payload.Data,
		}
		if !payload.Silent {
			msg.Notification = &messaging.Notification{Title: payload.Title, Body: payload.Body}
		}
		response, err := c.client.SendEachForMulticast(ctx, msg)
		if err != nil {
			failureCount += len(batch)
			continue
		}
		failureCount += response.FailureCount
		successCount += response.SuccessCount
	}
	return map[string]int{"success": successCount, "failure": failureCount}
}
