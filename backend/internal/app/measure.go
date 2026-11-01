package app

import "mapp/internal/measure"

// measureArea wires the Measure module (039): one register function per report, each in its own
// measure_<report>.go, so the lane that builds a report changes only that file (D-369).
func (w *wiring) measureArea() error {
	registry := measure.NewRegistry()
	registers := []func(*wiring, *measure.Registry) error{
		registerPlanBudget, registerApprovals, registerPurchaseRequests, registerPlanDetails, registerGsapLink,
		registerUserAccess, registerSummary, registerRevenue, registerFunnel,
	}
	for _, register := range registers {
		if err := register(w, registry); err != nil {
			return err
		}
	}
	return nil
}
