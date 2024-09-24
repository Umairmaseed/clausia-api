package db

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/goledgerdev/goprocess-api/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Notification struct defines a notification object
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"userId" json:"userId"`
	Type      string             `bson:"type" json:"type"`
	Message   string             `bson:"message" json:"message"`
	Metadata  map[string]string  `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Read      bool               `bson:"read" json:"read"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}

// NotificationService provides an interface to interact with notifications
type NotificationService struct {
	collection *mongo.Collection
}

// NewNotificationService returns a new NotificationService
func NewNotificationService(db *mongo.Database) *NotificationService {
	return &NotificationService{
		collection: db.Collection(notificationsCollection),
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, notif *[]Notification) (*mongo.InsertManyResult, error) {
	for i := range *notif {
		(*notif)[i].Timestamp = time.Now()
		(*notif)[i].Read = false
	}

	notifications := make([]interface{}, len(*notif))
	for i, n := range *notif {
		notifications[i] = n
	}

	result, err := s.collection.InsertMany(ctx, notifications)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *NotificationService) GetNotificationsByUser(ctx context.Context, userID string, limit int) ([]Notification, error) {
	var notifications []Notification
	opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notifID primitive.ObjectID) error {
	filter := bson.M{"_id": notifID}
	update := bson.M{"$set": bson.M{"read": true}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *NotificationService) MarkNotificationAsUnRead(ctx context.Context, notifID primitive.ObjectID) error {
	filter := bson.M{"_id": notifID}
	update := bson.M{"$set": bson.M{"read": false}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *NotificationService) DeleteNotification(ctx context.Context, notifID primitive.ObjectID) error {
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": notifID})
	return err
}

func (s *NotificationService) GetUnreadNotifications(ctx context.Context, userID string) ([]Notification, error) {
	var notifications []Notification
	cursor, err := s.collection.Find(ctx, bson.M{"userId": userID, "read": false})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *NotificationService) GetNotificationsByType(ctx context.Context, userID, notifType string) ([]Notification, error) {
	var notifications []Notification
	cursor, err := s.collection.Find(ctx, bson.M{"userId": userID, "type": notifType})
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func WatchForNotifications(mongodb *DB, server *websocket.WebSocketServer, ctx context.Context) {
	collection := mongodb.Database().Collection(notificationsCollection)

	changeStream, err := collection.Watch(ctx, mongo.Pipeline{})
	if err != nil {
		log.Fatalf("Error setting up change stream: %v", err)
	}
	defer changeStream.Close(ctx)

	for changeStream.Next(ctx) {
		var changeDoc bson.M
		if err := changeStream.Decode(&changeDoc); err != nil {
			log.Printf("Error decoding change event: %v", err)
			continue
		}

		fullDocument := changeDoc["fullDocument"].(bson.M)
		notification := Notification{
			ID:      fullDocument["_id"].(primitive.ObjectID),
			UserID:  fullDocument["userId"].(string),
			Type:    fullDocument["type"].(string),
			Message: fullDocument["message"].(string),
			Read:    fullDocument["read"].(bool),
		}

		if fullDocument["metadata"] != nil {
			notification.Metadata = make(map[string]string)
			for k, v := range fullDocument["metadata"].(bson.M) {
				notification.Metadata[k] = v.(string)
			}
		}

		NotifyUserOfChange(server, notification.UserID, notification)
	}
}

func NotifyUserOfChange(server *websocket.WebSocketServer, userID string, notification Notification) {
	notifData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshalling notification data for user %s: %v", userID, err)
		return
	}

	notificationMessage := websocket.NotificationMessage{
		UserID:  userID,
		Message: notifData,
	}
	server.Broadcast <- notificationMessage
}
