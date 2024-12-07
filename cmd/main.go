package main

import (
	"context"
	"fmt"
	"log"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/xifanyan/legalhold"
)

// Your auth provider implementation

func paginateUsers(ctx context.Context, gc *legalhold.MsGraphClient, userChan chan<- models.Userable) {
	defer close(userChan) // Ensure the channel is closed when done

	// Get all users with pagination
	requestBuilder := gc.Users()

	var pageSize int32 = 5
	// Create query options
	queryOptions := &users.UsersRequestBuilderGetQueryParameters{
		Top: &pageSize,
	}

	// Initial request to get the first page of users
	result, err := requestBuilder.Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: queryOptions,
	})
	if err != nil {
		log.Fatalf("Error getting users: %v", err)
	}

	// Send users to the channel
	sendUsersToChannel(result.GetValue(), userChan)

	// Check if there are more pages and paginate through them
	for result.GetOdataNextLink() != nil {
		nextLink := result.GetOdataNextLink()
		if nextLink == nil {
			break
		}
		requestBuilder = users.NewUsersRequestBuilder(*nextLink, gc.Adapter)
		result, err = requestBuilder.Get(ctx, nil)
		if err != nil {
			break
		}

		sendUsersToChannel(result.GetValue(), userChan)
	}
}

type User struct {
	Displayname *string
	Mail        *string
}

func sendUsersToChannel(users []models.Userable, userChan chan<- models.Userable) {
	fmt.Printf("SEND | %+v users to channel\n", users)

	for _, user := range users {
		if user != nil {
			userChan <- user
		}
	}
}

func main() {

	gc, err := legalhold.NewMsGraphClientBuilder().
		WithCertFile("C:/Users/pyan/lhn_msgraph_go.pfx").
		WithCertSecret("lhn_msgraph_go").
		WithTenantID("de62bccf-1ea0-44d1-a86e-e4918e21bbdc").
		WithClientID("dfa195d8-a531-4d58-ac9f-6053565a934a").
		Build()

	if err != nil {
		log.Fatalf("Error creating graph client: %v\n", err)
	}

	// Create a context
	ctx := context.Background()

	// Create a channel to receive users
	userChan := make(chan models.Userable)

	// Start pagination in a separate goroutine
	go paginateUsers(ctx, gc, userChan)

	// Process users received from the channel
	for user := range userChan {
		fmt.Printf("RECV | User: %s\n", *user.GetDisplayName())
	}
}
