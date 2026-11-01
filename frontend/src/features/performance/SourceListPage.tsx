import { Gauge, Plus } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/ui/Button/Button';
import { CanChange, SettingsViewOnlyNote } from '@/ui/CanChange/CanChange';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import { useCanChange } from '@/ui/useCanChange';
import { usePageTitle } from '@/ui/usePageTitle';
import tableStyles from '@/ui/settings-table.module.css';
import { useFormatters } from '@/format/useFormatters';
import { usePerformanceSources } from './api';
import type { PerformanceSource } from './api';
import { SourceForm } from './SourceForm';
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
type OpenForm = { kind: 'new' } | { kind: 'edit'; source: PerformanceSource };

export function SourceListPage() {
  usePageTitle(PERFORMANCE_DATA);
  const sources = usePerformanceSources();
  const [form, setForm] = useState<OpenForm | null>(null);
  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Settings', to: '/settings' }, { label: PERFORMANCE_DATA }]}
        title={PERFORMANCE_DATA}
        description={SETTINGS_DESCRIPTION}
        actions={
          <CanChange permission="performance.import" otherwise={<SettingsViewOnlyNote />}>
            <Button icon={<Plus aria-hidden />} onClick={() => setForm({ kind: 'new' })}>
              New data source
            </Button>
          </CanChange>
        }
      />
      <Card>
        {sources.isError ? (
          <ErrorState error={sources.error} onRetry={() => void sources.refetch()} />
        ) : sources.data ? (
          <SourceTable
            items={sources.data.items}
            onEdit={(source) => setForm({ kind: 'edit', source })}
          />
        ) : (
          <Skeleton lines={4} />
        )}
      </Card>
      {form && (
        <SourceForm
          source={form.kind === 'edit' ? form.source : undefined}
          onClose={() => setForm(null)}
        />
      )}
    </>
  );
}

function SourceTable({
  items,
  onEdit,
}: {
  items: PerformanceSource[];
  onEdit: (s: PerformanceSource) => void;
}) {
  const { date } = useFormatters();
  const canImport = useCanChange('performance.import');
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
          <th scope="col">
            <span className="visually-hidden">Actions</span>
          </th>
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
            <td>
              {canImport && (
                <Button variant="tertiary" onClick={() => onEdit(source)}>
                  Edit
                </Button>
              )}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
