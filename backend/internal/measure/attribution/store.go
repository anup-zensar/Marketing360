package attribution

import "context"

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
	Write(ctx context.Context, fn func(ctx context.Context) error) error
}

// Store keeps the designated model; Settings gives DefaultSettings until it was first changed.
type Store interface {
	Settings(ctx context.Context) (Settings, error)
	PutSettings(ctx context.Context, s Settings) error
}
