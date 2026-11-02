package revenue

import (
	"testing"

	"mapp/internal/platform/money"
)

func TestPerformanceOf_RoundsToAWholePerCent_FR023(t *testing.T) {
	t.Parallel()

	// 14,939 / 370,000 = 4.04...%, which must round to the nearest whole per cent (049-C15).
	got := performanceOf(money.FromCents(14_939), money.FromCents(370_000), true)

	if got.Value == nil || *got.Value != 4 {
		t.Errorf("performance = %v, want 4 (rounded)", got.Value)
	}
}
