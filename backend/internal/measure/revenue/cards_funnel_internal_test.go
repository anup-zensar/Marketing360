package revenue

import "testing"

func TestRatioCard_RoundsToAWholePerCent_049C15(t *testing.T) {
	t.Parallel()

	got := ratioCard(149, 370_000, true)

	if got.Value == nil {
		t.Fatalf("performance = %+v, want a value: a goal is in view", got)
	}
	// 149 / 370,000 = 0.0402...%, rounded to the nearest whole per cent (049-C15).
	if *got.Value != 0 {
		t.Errorf("performance = %v, want 0 (rounded)", *got.Value)
	}

	got = ratioCard(14_939, 370_000, true)
	if got.Value == nil || *got.Value != 4 {
		t.Errorf("performance = %v, want 4 (14,939 / 370,000 = 4.04%%, rounded)", got.Value)
	}
}
