package stores

import "mapp/internal/measure"

// MeasureStore keeps the saved report views (feature 039).
type MeasureStore interface {
	measure.Store
}
