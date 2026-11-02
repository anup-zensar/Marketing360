package stores

import "mapp/internal/measure/revenue"

// RevenueStore is feature 049's one new read: the counted facts of the activities in a report's view.
type RevenueStore interface {
	revenue.Facts
}
