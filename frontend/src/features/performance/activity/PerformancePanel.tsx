import { Card } from '@/ui/Card/Card';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import type { StatusTone } from '@/ui/StatusChip/StatusChip';
import { formatWhole } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { usePerformancePanel } from './api';
import type {
  MetricFigure,
  PanelSource,
  PerformancePanel as Panel,
  RevenueFigure,
  StageFigure,
} from './api';
import styles from './DataSourcesArea.module.css';

const GATE = 'Approve the activity to see performance data.';

/** Why a figure is empty (FR-032): the API's stated reason, never inferred and never a sample value. */
const REASON_WORDS: Record<string, string> = {
  noSource: 'No active data source supplies this figure.',
  noGoal: 'No goal is set for this figure.',
  noData: 'No data has arrived yet.',
  stale: 'The data source is stale.',
  failed: 'The last load failed.',
  notApproved: GATE,
};

const STAGE_WORDS: Record<string, string> = { inq: 'INQ', mql: 'MQL' };
const METRIC_WORDS: Record<string, string> = {
  sessions: 'Sessions',
  visits: 'Visits',
  engagementTouches: 'Engagement touches',
};
const FRESHNESS_TONE: Record<string, StatusTone> = {
  current: 'positive',
  stale: 'review',
  failed: 'problem',
  inactive: 'neutral',
};

/** The activity's Performance area (046 FR-027, FR-029): figures gated until it counts as approved. */
export function PerformancePanel({ activityId }: { activityId: string }) {
  const panel = usePerformancePanel(activityId);
  return (
    <Card title="Performance">
      {panel.isError ? (
        <ErrorState error={panel.error} onRetry={() => void panel.refetch()} />
      ) : panel.data ? (
        <PanelBody panel={panel.data} />
      ) : (
        <Skeleton lines={4} />
      )}
    </Card>
  );
}

function PanelBody({ panel }: { panel: Panel }) {
  if (panel.state === 'notApproved') {
    return <p className={styles.notice}>{GATE}</p>;
  }
  return (
    <div className={styles.stack}>
      <StageTable stages={panel.stages} revenue={panel.revenue} />
      {panel.webMetrics.length > 0 && <WebMetrics metrics={panel.webMetrics} />}
      {panel.sources.length > 0 && <Sources sources={panel.sources} />}
    </div>
  );
}

function StageTable({ stages, revenue }: { stages: StageFigure[]; revenue: RevenueFigure }) {
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th scope="col">Stage</th>
          <th scope="col">Goal</th>
          <th scope="col">Actual</th>
          <th scope="col">Conversion</th>
        </tr>
      </thead>
      <tbody>
        {stages.map((stage) => (
          <tr key={stage.stage}>
            <th scope="row">{STAGE_WORDS[stage.stage] ?? stage.stage}</th>
            {stage.reason ? (
              <td colSpan={3} className={styles.reason}>
                {REASON_WORDS[stage.reason] ?? stage.reason}
              </td>
            ) : (
              <>
                <td className={styles.numeric}>
                  {stage.goal === null ? '—' : formatWhole(stage.goal)}
                </td>
                <td className={styles.numeric}>
                  {stage.actual === null ? '—' : formatWhole(stage.actual)}
                </td>
                <td className={styles.numeric}>
                  {stage.conversion === null ? '—' : `${stage.conversion.toFixed(1)} %`}
                </td>
              </>
            )}
          </tr>
        ))}
        <RevenueRow revenue={revenue} />
      </tbody>
    </table>
  );
}

function RevenueRow({ revenue }: { revenue: RevenueFigure }) {
  const { money } = useFormatters();
  if (revenue.reason) {
    return (
      <tr>
        <th scope="row">Revenue</th>
        <td colSpan={3} className={styles.reason}>
          {REASON_WORDS[revenue.reason] ?? revenue.reason}
        </td>
      </tr>
    );
  }
  return (
    <tr>
      <th scope="row">Revenue</th>
      <td className={styles.numeric}>{revenue.goal ? money(revenue.goal) : '—'}</td>
      <td className={styles.numeric}>
        {revenue.actual ? money(revenue.actual) : '—'}
        {revenue.excludedCount > 0 && (
          <span className={styles.reason}> · {revenue.excludedCount} excluded without a rate</span>
        )}
      </td>
      <td className={styles.numeric}>—</td>
    </tr>
  );
}

function WebMetrics({ metrics }: { metrics: MetricFigure[] }) {
  return (
    <div>
      {metrics.map((metric) => (
        <p key={metric.metric} className={styles.notice}>
          {METRIC_WORDS[metric.metric] ?? metric.metric}:{' '}
          {metric.months.map((m) => `${m.month} · ${formatWhole(m.value)}`).join(', ')}
        </p>
      ))}
    </div>
  );
}

function Sources({ sources }: { sources: PanelSource[] }) {
  const { date } = useFormatters();
  return (
    <p className={styles.notice}>
      {sources.map((source) => (
        <span key={source.sourceId} className={styles.alias}>
          {source.sourceName}{' '}
          <StatusChip tone={FRESHNESS_TONE[source.freshness] ?? 'neutral'}>
            {source.freshness}
          </StatusChip>
          {source.dataThrough && <> · through {date(source.dataThrough)}</>}
        </span>
      ))}
    </p>
  );
}
