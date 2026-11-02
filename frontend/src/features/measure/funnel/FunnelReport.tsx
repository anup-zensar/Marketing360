import { useState } from 'react';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { RankedList, ReportFrame, useRankedState } from '../shared';
import type { RankedColumn, ReportAddress, ReportFrameData } from '../shared';
import { useFunnelReport, useFunnelRanked } from '../revenue/api';
import type { FunnelFigures, FunnelRankedParams } from '../revenue/api';
import { CategorySelect } from '../revenue/CategorySelect';
import { FunnelCards } from './FunnelCards';
import { FunnelChart } from './FunnelChart';
import { StageSelect } from './StageSelect';
import styles from './Funnel.module.css';

const LEVELS = ['plan', 'campaign', 'initiative', 'activity'];
const DEFAULT_SORT = 'projected';

/** One funnel stage against goal (feature 049 US2): cards, series and 040's ranked list, in 039's frame. */
export function FunnelReport() {
  return (
    <ReportFrame report="funnel" period show="campaign">
      {(frame, address) => <FunnelBody frame={frame} address={address} />}
    </ReportFrame>
  );
}

function FunnelBody({ frame, address }: { frame: ReportFrameData; address: ReportAddress }) {
  const [category, setCategory] = useState('');
  const [stage, setStage] = useState('');
  const report = useFunnelReport(address, {
    category: category || undefined,
    stage: isStage(stage) ? stage : undefined,
  });
  const chosenStage = stage || report.data?.stage || 'inq';
  const { state } = useRankedState(DEFAULT_SORT);
  const ranked = useFunnelRanked(address, rankedParams(state, category, chosenStage));
  const show = address.show ?? 'campaign';
  const whole = (value: number | null) => (value === null ? '—' : value.toLocaleString());
  const columns: RankedColumn<FunnelFigures>[] = [
    { key: 'goal', label: 'Goal', value: (f) => whole(f.goal) },
    { key: 'actual', label: 'Actual', value: (f) => whole(f.actual) },
    { key: 'projected', label: 'Projected', value: (f) => whole(f.projected) },
    {
      key: 'performance',
      label: 'Performance',
      value: (f) => (f.performance !== null ? `${f.performance}%` : '—'),
    },
    {
      key: 'conversion',
      label: 'Conversion',
      value: (f) => (f.conversion !== null ? `${f.conversion}%` : '—'),
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
        <StageSelect
          stage={chosenStage}
          available={report.data?.availableStages ?? [chosenStage]}
          onChange={setStage}
        />
        <CategorySelect value={category} onChange={setCategory} />
      </div>
      {report.data ? (
        <>
          <FunnelCards stage={report.data.stage} cards={report.data.cards} />
          <FunnelChart report={report.data} />
          {report.data.uncoveredCount > 0 && (
            <p>
              {report.data.uncoveredCount} plan{report.data.uncoveredCount === 1 ? '' : 's'} use a
              model that does not allow this stage.
            </p>
          )}
        </>
      ) : report.isError ? (
        <ErrorState error={report.error} onRetry={() => void report.refetch()} />
      ) : (
        <div aria-label="Loading the figures">
          <Skeleton lines={4} height="24px" />
        </div>
      )}
      <RankedList
        report="funnel"
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
  stage: string,
): FunnelRankedParams {
  const sorts = ['actual', 'projected'] as const;
  return {
    sort: sorts.find((s) => s === state.sort) ?? 'projected',
    order: state.order,
    size: state.size,
    page: state.page,
    ...(state.group !== 'none' && { group: state.group }),
    ...(state.q && { q: state.q }),
    ...(state.drill && { drill: state.drill }),
    ...(category && { category }),
    ...(isStage(stage) && { stage }),
  };
}

function isStage(value: string): value is 'inq' | 'mql' {
  return value === 'inq' || value === 'mql';
}
