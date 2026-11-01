package plandetails

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

// The All work columns the report needs: the completeness of every record and the activities' goals.
const (
	columnCompleteness = "completeness"
	columnGoals        = "inq"
)

type Service struct {
	tx     Transactor
	scoped Scoped
}

func NewService(tx Transactor, scoped Scoped) *Service {
	return &Service{tx: tx, scoped: scoped}
}

// Report works out the level f.Show (Activity by default) of the records matching f, in one Read.
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter) (Report, error) {
	var r Report
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		m, err := s.scoped.Matched(ctx, a, f, []string{columnCompleteness, columnGoals})
		if err != nil {
			return err
		}
		attributes, err := s.scoped.Attributes(ctx)
		if err != nil {
			return err
		}
		v := newView(m, levelOf(f.Show), f.Period)
		rows := v.rows(attributes)
		r = Report{Show: v.level, Currency: m.Currency.Code, Cards: v.cards(), BuiltIn: v.builtIn(), Attributes: rows,
			AttributesSetUp: len(rows) > 0}
		return nil
	})
	if err != nil {
		return Report{}, err
	}
	return r, nil
}

func levelOf(show string) Level {
	if show == "" {
		return LevelActivity
	}
	return Level(show)
}
