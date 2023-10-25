package account_test

import (
	"context"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/auth"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/account"
	"github.com/google/uuid"
	"golang.org/x/oauth2/clientcredentials"
	"testing"
)

func Test(t *testing.T) {
	oauthClient := auth.NewOAuthBasedClient(context.TODO(), clientcredentials.Config{})

	c, _ := account.NewClient("https://api.dynatrace.com", account.WithHTTPClient(oauthClient))

	resp, _ := c.UsersControllerGetUsers(context.TODO(), uuid.New().String(), nil)

	res, _ := account.ParseUsersControllerGetUsersResponse(resp)
	users := res.JSON200.Items
	_ = users
}
