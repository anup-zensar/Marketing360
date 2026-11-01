import { useFormatters } from '@/format/useFormatters';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { RankedList, ReportFrame, useRankedState } from '../shared';
import type { RankedColumn, ReportAddress, ReportFrameData } from '../shared';
import { usePlanBudget, usePlanBudgetRanked } from './api';
import type { PlanBudgetFigures, RankedParams } from './api';
import { BudgetCards } from './BudgetCards';
import { PacingChart } from './PacingChart';
import styles from './PlanBudget.module.css';

const LEVELS = ['plan', 'campaign', 'initiative', 'activity'];
const DEFAULT_SORT = 'planned';

/** Plan budget (feature 040): spend pacing by period and the ranked list, in 039's frame. */
export function PlanBudgetReport() {
  return (
    <ReportFrame report="planBudget" period show="campaign">
      {(frame, address) => <PlanBudgetBody frame={frame} address={address} />}
    </ReportFrame>
  );
}

function PlanBudgetBody({ frame, address }: { frame: ReportFrameData; address: ReportAddress }) {
  const { money } = useFormatters();
  const report = usePlanBudget(address);
  const { state } = useRankedState(DEFAULT_SORT);
  const ranked = usePlanBudgetRanked(address, rankedParams(state));
  const show = address.show ?? 'campaign';
  const shown = (value: PlanBudgetFigures['budget']) => (value ? money(value) : '—');
  const columns: RankedColumn<PlanBudgetFigures>[] = [
    { key: 'budget', label: 'Budget', value: (f) => shown(f.budget) },
    { key: 'planned', label: 'Planned cost', value: (f) => shown(f.planned) },
    { key: 'committed', label: 'Committed', value: (f) => shown(f.committed) },
    { key: 'actual', label: 'Actual', value: (f) => shown(f.actual) },
    { key: 'leftToPlan', label: 'Left to plan', value: (f) => shown(f.leftToPlan) },
    { key: 'unspent', label: 'Unspent budget', value: (f) => shown(f.unspent) },
  ];
  const groups = [
    { value: 'none', label: 'None' },
    { value: 'plan', label: 'Plan' },
    { value: 'owner', label: 'Owner' },
    ...frame.choices.attributeGroups
      .filter((g) => LEVELS.indexOf(g.level) <= LEVELS.indexOf(show))
      .map((g) => ({ value: `attr:${g.attributeId}`, label: g.label })),
  ];
  return (
    <div className={styles.body}>
      {report.data ? (
        <>
          <BudgetCards cards={report.data.cards} />
          <PacingChart report={report.data} />
        </>
      ) : report.isError ? (
        <ErrorState error={report.error} onRetry={() => void report.refetch()} />
      ) : (
        <div aria-label="Loading the figures">
          <Skeleton lines={4} height="24px" />
        </div>
      )}
      <RankedList
        report="planBudget"
        data={ranked.data}
        error={ranked.isError ? ranked.error : null}
        onRetry={() => void ranked.refetch()}
        columns={columns}
        groups={groups}
        defaultSort={DEFAULT_SORT}
        flag={(f) => (f.overAllocated ? 'Over-allocated' : undefined)}
      />
    </div>
  );
}

function rankedParams(state: ReturnType<typeof useRankedState>['state']): RankedParams {
  const sorts = ['budget', 'planned', 'committed', 'actual', 'leftToPlan', 'unspent'] as const;
  return {
    sort: sorts.find((s) => s === state.sort) ?? 'planned',
    order: state.order,
    size: state.size,
    page: state.page,
    ...(state.group !== 'none' && { group: state.group }),
    ...(state.q && { q: state.q }),
    ...(state.drill && { drill: state.drill }),
  };
}
