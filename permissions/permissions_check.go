package permissions

import (
	"context"
	"slices"

	"strings"

	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"

	"github.com/ONSdigital/log.go/v2/log"
)

func CheckIsAdmin(ctx context.Context, token string, authorisation auth.Middleware) (bool, error) {
	userToken := strings.ReplaceAll(token, "Bearer ", "")
	entityData, err := authorisation.Parse(userToken)
	if err != nil {
		log.Error(ctx, "Cannot parse user JWT token", err)
		return false, err
	}

	if slices.Contains(entityData.Groups, "role-admin") {
		return true, nil
	}

	return false, nil
}
