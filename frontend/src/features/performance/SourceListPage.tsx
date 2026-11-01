import { Gauge } from 'lucide-react';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import { usePageTitle } from '@/ui/usePageTitle';
import tableStyles from '@/ui/settings-table.module.css';
import { useFormatters } from '@/format/useFormatters';
import { usePerformanceSources } from './api';
import type { PerformanceSource } from './api';
import {
  CADENCE_WORDS,
  CATEGORY_WORDS,
  EMPTY_TEXT,
  EMPTY_TITLE,
  FRESHNESS_TONES,
  FRESHNESS_WORDS,
  NEVER_LOADED,
  PERFORMANCE_DATA,
  SETTINGS_DESCRIPTION,
} from './performance-text';

/**
 * Settings › Performance data (feature 046, FR-001, FR-005): every data source with the category it
 * belongs to, how often it delivers, and the freshness derived from its last load.
 */
export function SourceListPage() {
  usePageTitle(PERFORMANCE_DATA);
  const sources = usePerformanceSources();
  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Settings', to: '/settings' }, { label: PERFORMANCE_DATA }]}
        title={PERFORMANCE_DATA}
        description={SETTINGS_DESCRIPTION}
      />
      <Card>
        {sources.isError ? (
          <ErrorState error={sources.error} onRetry={() => void sources.refetch()} />
        ) : sources.data ? (
          <SourceTable items={sources.data.items} />
        ) : (
          <Skeleton lines={4} />
        )}
      </Card>
    </>
  );
}

function SourceTable({ items }: { items: PerformanceSource[] }) {
  const { date } = useFormatters();
  if (items.length === 0) {
    return <EmptyState icon={<Gauge aria-hidden />} title={EMPTY_TITLE} text={EMPTY_TEXT} />;
  }
  return (
    <table className={tableStyles.table}>
      <thead>
        <tr>
          <th scope="col">Name</th>
          <th scope="col">Category</th>
          <th scope="col">Delivers</th>
          <th scope="col">Last loaded</th>
          <th scope="col">Freshness</th>
        </tr>
      </thead>
      <tbody>
        {items.map((source) => (
          <tr key={source.id}>
            <td>{source.name}</td>
            <td>{CATEGORY_WORDS[source.category]}</td>
            <td>{CADENCE_WORDS[source.cadence]}</td>
            <td>{source.lastLoadedAt ? date(source.lastLoadedAt) : NEVER_LOADED}</td>
            <td>
              <StatusChip tone={FRESHNESS_TONES[source.freshness]}>
                {FRESHNESS_WORDS[source.freshness]}
              </StatusChip>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
