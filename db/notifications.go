package db

import (
	"context"
	"time"

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

func (s *NotificationService) CreateNotification(ctx context.Context, notif *Notification) (*mongo.InsertOneResult, error) {
	notif.Timestamp = time.Now()
	result, err := s.collection.InsertOne(ctx, notif)
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
