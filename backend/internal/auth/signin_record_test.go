package auth

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func lastSignIns(t *testing.T, f *fixture) map[string]time.Time {
	t.Helper()
	last, err := newFakeStore(f.db).LastSignIns(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return last
}

func TestSignIn_RecordsLastSignIn_FR009(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	f.signIn(t, priya.Email, "")
	f.clock.Advance(2 * time.Hour)
	f.signIn(t, daniel.Email, "")
	f.signIn(t, priya.Email, "")

	want := map[string]time.Time{priya.ID: start.Add(2 * time.Hour), daniel.ID: start.Add(2 * time.Hour)}
	if diff := cmp.Diff(want, lastSignIns(t, f)); diff != "" {
		t.Errorf("last sign-ins (-want +got):\n%s", diff)
	}
}

func TestSignIn_RefusedRecordsNothing_FR009(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.deactivate(t, mei)

	_, _, wrong := f.service.SignIn(t.Context(), SignInAttempt{Email: priya.Email, Password: "wrong", Client: client})
	_, _, inactive := f.service.SignIn(t.Context(), SignInAttempt{Email: mei.Email, Password: demoPassword})
	for range emailLimit {
		_, _, _ = f.service.SignIn(t.Context(), SignInAttempt{Email: daniel.Email, Password: "wrong", Client: client})
	}
	_, _, locked := f.service.SignIn(t.Context(), SignInAttempt{Email: daniel.Email, Password: demoPassword,
		Client: otherClient})

	if wrong == nil || inactive == nil || locked == nil {
		t.Fatalf("refusals = %v, %v, %v; want three errors", wrong, inactive, locked)
	}
	if last := lastSignIns(t, f); len(last) != 0 {
		t.Errorf("last sign-ins after refusals = %v, want none", last)
	}
}
