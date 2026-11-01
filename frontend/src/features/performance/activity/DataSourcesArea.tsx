import { useState } from 'react';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { useToast } from '@/ui/Toast/useToast';
import { useCanChange } from '@/ui/useCanChange';
import { Detail, DetailsList } from '@/features/line-items';
import { useFormatters } from '@/format/useFormatters';
import { useDataSourcesArea, useDeleteAlias, useDeleteIntegrationId } from './api';
import type { DataSourcesArea as Area, ExternalNameAlias, SourceOfFacts } from './api';
import { KeyDialog } from './KeyDialog';
import styles from './DataSourcesArea.module.css';

const NO_SOURCES =
  'No data source delivers for this activity yet. Register an integration ID, or upload with its activity ID as the key.';

/** The activity's Data sources area (046 FR-031): its matching keys, and what each source delivered. */
export function DataSourcesArea({ activityId }: { activityId: string }) {
  const area = useDataSourcesArea(activityId);
  return (
    <Card title="Data sources">
      {area.isError ? (
        <ErrorState error={area.error} onRetry={() => void area.refetch()} />
      ) : area.data ? (
        <AreaBody activityId={activityId} area={area.data} />
      ) : (
        <Skeleton lines={4} />
      )}
    </Card>
  );
}

function AreaBody({ activityId, area }: { activityId: string; area: Area }) {
  const canLink = useCanChange('performance.link');
  const [adding, setAdding] = useState(false);
  return (
    <div className={styles.stack}>
      <DetailsList>
        <Detail label="External name">{area.externalName || null}</Detail>
        <Detail label="Previous names">
          <Aliases activityId={activityId} aliases={area.aliases} canLink={canLink} />
        </Detail>
        <Detail label="Is integrated">{area.integrated ? 'Yes' : 'No'}</Detail>
        <Detail label="Goal entered">{area.goalEntered ? 'Yes' : 'No'}</Detail>
        <Detail label="Media codes" numeric>
          {String(area.mediaCodeCount)}
        </Detail>
      </DetailsList>
      {area.sources.length === 0 ? (
        <p className={styles.notice}>{NO_SOURCES}</p>
      ) : (
        <SourceRows activityId={activityId} sources={area.sources} canLink={canLink} />
      )}
      {canLink && (
        <div>
          <Button variant="secondary" onClick={() => setAdding(true)}>
            Register integration ID
          </Button>
        </div>
      )}
      {adding && <KeyDialog activityId={activityId} onClose={() => setAdding(false)} />}
    </div>
  );
}

function Aliases(props: { activityId: string; aliases: ExternalNameAlias[]; canLink: boolean }) {
  const { show } = useToast();
  const deleteAlias = useDeleteAlias(props.activityId);
  const [removing, setRemoving] = useState<ExternalNameAlias | null>(null);
  if (props.aliases.length === 0) {
    return null;
  }
  return (
    <span className={styles.aliases}>
      {props.aliases.map((alias) => (
        <span key={alias.id} className={styles.alias}>
          {alias.name}
          {props.canLink && (
            <button
              type="button"
              className={styles.remove}
              aria-label={`Remove the previous name ${alias.name}`}
              onClick={() => setRemoving(alias)}
            >
              ×
            </button>
          )}
        </span>
      ))}
      {removing && (
        <ConfirmDialog
          open
          onOpenChange={(open) => {
            if (!open) setRemoving(null);
          }}
          title="Remove this previous name?"
          confirmLabel="Remove name"
          tone="danger"
          busy={deleteAlias.isPending}
          onConfirm={() => {
            deleteAlias.mutate(removing.id, {
              onSuccess: () => {
                show({
                  message: `'${removing.name}' removed — rows carrying it will stop matching`,
                });
                setRemoving(null);
              },
            });
          }}
        >
          Rows that still carry ‘{removing.name}’ will stop matching this activity.
        </ConfirmDialog>
      )}
    </span>
  );
}

function SourceRows(props: { activityId: string; sources: SourceOfFacts[]; canLink: boolean }) {
  const { show } = useToast();
  const { date } = useFormatters();
  const deleteKey = useDeleteIntegrationId(props.activityId);
  const [removing, setRemoving] = useState<SourceOfFacts | null>(null);
  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th scope="col">Source</th>
          <th scope="col">Integration ID</th>
          <th scope="col">Facts</th>
          <th scope="col">Last fact</th>
          {props.canLink && (
            <th scope="col">
              <span className="visually-hidden">Actions</span>
            </th>
          )}
        </tr>
      </thead>
      <tbody>
        {props.sources.map((source) => (
          <tr key={source.sourceId}>
            <td>{source.sourceName || source.sourceId}</td>
            <td>{source.integrationId || '—'}</td>
            <td className={styles.numeric}>{source.factCount}</td>
            <td>{source.lastFactDate ? date(source.lastFactDate) : '—'}</td>
            {props.canLink && (
              <td>
                {source.integrationId && (
                  <Button variant="tertiary" onClick={() => setRemoving(source)}>
                    Remove ID
                  </Button>
                )}
              </td>
            )}
          </tr>
        ))}
        {removing && (
          <ConfirmDialog
            open
            onOpenChange={(open) => {
              if (!open) setRemoving(null);
            }}
            title="Remove this integration ID?"
            confirmLabel="Remove ID"
            tone="danger"
            busy={deleteKey.isPending}
            onConfirm={() => {
              deleteKey.mutate(removing.sourceId, {
                onSuccess: () => {
                  show({
                    message: 'Integration ID removed — that source’s rows will stop matching here',
                  });
                  setRemoving(null);
                },
              });
            }}
          >
            Rows from {removing.sourceName || removing.sourceId} carrying ‘{removing.integrationId}’
            will stop matching this activity.
          </ConfirmDialog>
        )}
      </tbody>
    </table>
  );
}
