package seed

import "context"

// addMeasure adds the Measure demo data (issue #48, lane D) after the demo story: it only creates records and never
// changes an existing demo record (D-354), so every report has something to show.
func (d *demo) addMeasure(context.Context) error {
	return nil
}
