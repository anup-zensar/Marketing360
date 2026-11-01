package measure

import "context"

type Store interface {
	ListViews(ctx context.Context, personID, report string) ([]StoredView, error)
	GetView(ctx context.Context, id string) (StoredView, bool, error)
	PutView(ctx context.Context, v StoredView) error
	DeleteView(ctx context.Context, id string) error
}
