package stores

import "mapp/internal/measure/attribution"

// AttributionStore keeps the designated attribution model (feature 050).
type AttributionStore interface {
	attribution.Store
}
