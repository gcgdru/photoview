package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.63

import (
	"context"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/notification"
)

// Notification is the resolver for the notification field.
func (r *subscriptionResolver) Notification(ctx context.Context) (<-chan *models.Notification, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, auth.ErrUnauthorized
	}

	notificationChannel := make(chan *models.Notification, 1)

	listenerID := notification.RegisterListener(user, notificationChannel)

	go func() {
		<-ctx.Done()
		notification.DeregisterListener(listenerID)
	}()

	return notificationChannel, nil
}

// Subscription returns api.SubscriptionResolver implementation.
func (r *Resolver) Subscription() api.SubscriptionResolver { return &subscriptionResolver{r} }

type subscriptionResolver struct{ *Resolver }
