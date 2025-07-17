package app

import (
	"context"
	"errors"
	"fmt"
	"pr-mail/pkg/middleware"
	"strconv"
	"strings"
)

// GetAdminIDFromContext retrieves the admin ID from context
func GetAdminIDFromContext(ctx context.Context) (int64, error) {
	adminID, ok := ctx.Value(middleware.AdminIDKey).(int64)
	if !ok {
		return 0, errors.New("admin ID not found in context")
	}
	return adminID, nil
}

// GetAdminUsernameFromContext retrieves the admin username from context
func GetAdminUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value(middleware.AdminNameKey).(string)
	if !ok {
		return "", errors.New("admin username not found in context")
	}
	return username, nil
}

func ParsePRLink(link string) (owner, repo string, prNumber int, err error) {
	// Example: https://github.com/srz-vishnu/E-cart-FE/pull/10
	parts := strings.Split(link, "/")
	if len(parts) < 7 {
		return "", "", 0, fmt.Errorf("invalid PR link")
	}
	owner = parts[3]
	repo = parts[4]
	prNumber, err = strconv.Atoi(parts[6])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid PR number in link")
	}
	return owner, repo, prNumber, nil
}
