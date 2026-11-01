import { useState } from 'react';
import { useFormatters } from '@/format/useFormatters';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { RankedList, ReportFrame, useRankedState } from '../shared';
import type { RankedColumn, ReportAddress, ReportFrameData } from '../shared';
import { useRevenueReport, useRevenueRanked } from './api';
import type { RevenueFigures, RevenueRankedParams } from './api';
import { CategorySelect } from './CategorySelect';
import { RevenueCards } from './RevenueCards';
import { RevenueChart } from './RevenueChart';
import { reasonText } from './reasons';
import styles from './Revenue.module.css';

const LEVELS = ['plan', 'campaign', 'initiative', 'activity'];
const DEFAULT_SORT = 'projectedRevenue';

/** Revenue against goal (feature 049 US1): cards, the period series and 040's ranked list, in 039's frame. */
export function RevenueReport() {
  return (
    <ReportFrame report="revenue" period show="campaign">
      {(frame, address) => <RevenueBody frame={frame} address={address} />}
    </ReportFrame>
  );
}

function RevenueBody({ frame, address }: { frame: ReportFrameData; address: ReportAddress }) {
  const { money } = useFormatters();
  const [category, setCategory] = useState('');
  const report = useRevenueReport(address, { category: category || undefined });
  const { state } = useRankedState(DEFAULT_SORT);
  const ranked = useRevenueRanked(address, rankedParams(state, category));
  const show = address.show ?? 'campaign';
  const columns: RankedColumn<RevenueFigures>[] = [
    { key: 'goal', label: 'Goal', value: (f) => (f.goal ? money(f.goal) : '—') },
    { key: 'actual', label: 'Actual', value: (f) => (f.actual ? money(f.actual) : '—') },
    {
      key: 'projected',
      label: 'Projected',
      value: (f) => (f.projected ? money(f.projected) : '—'),
    },
    {
      key: 'performance',
      label: 'Performance',
      value: (f) => (f.performance !== null ? `${f.performance}%` : '—'),
    },
    {
      key: 'actualCost',
      label: 'Actual cost',
      value: (f) => (f.actualCost ? money(f.actualCost) : reasonText(f.actualCostReason)),
    },
    {
      key: 'roi',
      label: 'ROI',
      value: (f) => (f.roi !== null ? `${f.roi}%` : reasonText(f.roiReason)),
    },
  ];
  const groups = [
    { value: 'none', label: 'None' },
    { value: 'plan', label: 'Plan' },
    { value: 'owner', label: 'Owner' },
    ...frame.choices.attributeGroups
      .filter((g) => LEVELS.includes(g.level) && LEVELS.indexOf(g.level) <= LEVELS.indexOf(show))
      .map((g) => ({ value: `attr:${g.attributeId}`, label: g.label })),
  ];
  return (
    <div className={styles.body}>
      <div className={styles.controls}>
        <CategorySelect value={category} onChange={setCategory} />
      </div>
      {report.data ? (
        <>
          <RevenueCards cards={report.data.cards} />
          <RevenueChart report={report.data} />
        </>
      ) : report.isError ? (
        <ErrorState error={report.error} onRetry={() => void report.refetch()} />
      ) : (
        <div aria-label="Loading the figures">
          <Skeleton lines={4} height="24px" />
        </div>
      )}
      <RankedList
        report="revenue"
        data={ranked.data}
        error={ranked.isError ? ranked.error : null}
        onRetry={() => void ranked.refetch()}
        columns={columns}
        groups={groups}
        defaultSort={DEFAULT_SORT}
      />
    </div>
  );
}

function rankedParams(
  state: ReturnType<typeof useRankedState>['state'],
  category: string,
): RevenueRankedParams {
  const sorts = ['actualRevenue', 'projectedRevenue'] as const;
  return {
    sort: sorts.find((s) => s === state.sort) ?? 'projectedRevenue',
    order: state.order,
    size: state.size,
    page: state.page,
    ...(state.group !== 'none' && { group: state.group }),
    ...(state.q && { q: state.q }),
    ...(state.drill && { drill: state.drill }),
    ...(category && { category }),
  };
}
