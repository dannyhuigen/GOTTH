package session

import (
	"context"
	"goth/internal/store"
)

type Session struct {
	IsDemo             bool
	CurrentUser        *store.GoogleUser
	CurrentWorkspace   *any
	AllWorkspaces      []*any
	StripeSubscription *any
}

func GetSessionFromCtx(ctx context.Context) *Session {
	var value = ctx.Value("session")
	if value == nil {
		return nil
	}
	return value.(*Session)
}

func GetUserNameFromCtx(ctx context.Context) string {
	var session = GetSessionFromCtx(ctx)
	if session != nil && session.CurrentUser != nil {
		return session.CurrentUser.Name
	}
	return "Demo user"
}

func GetProfilePicFromCtx(ctx context.Context) string {
	var session = GetSessionFromCtx(ctx)
	if session != nil && session.CurrentUser != nil {
		return session.CurrentUser.Picture
	}
	return "Demo user"
}
