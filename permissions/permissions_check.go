package permissions

import (
	"context"
	"fmt"
	"slices"

	"strings"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
)

func CheckIsAdmin(ctx context.Context, token string, authorisation auth.Middleware) (bool, error) {
	userToken := strings.ReplaceAll(token, "Bearer ", "")
	entityData, err := authorisation.Parse(userToken)
	if err != nil {
		return false, fmt.Errorf("check admin: failed to parse user JWT token: %w", err)
	}

	if slices.Contains(entityData.Groups, "role-admin") {
		return true, nil
	}

	return false, nil
}
