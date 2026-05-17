package configs

import (
	"context"
	"zenith-on-stage/internal/model"
)

type AuthenticationStore interface {
	SetState(ctx context.Context, state string) error
	GetState(ctx context.Context, state string) (string, error)
	DeleteState(ctx context.Context, state string) error
}

// SessionStore defines the contract for session management
type SessionStore interface {
	Set(ctx context.Context, sessionID string, data model.SessionData) error
	Get(ctx context.Context, sessionID string) (*model.SessionData, error)
	Delete(ctx context.Context, sessionID string) error
}
