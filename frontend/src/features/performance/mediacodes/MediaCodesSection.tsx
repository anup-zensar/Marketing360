import { useState } from 'react';
import { asApiError } from '@/api/errors';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import { useToast } from '@/ui/Toast/useToast';
import { useCanChange } from '@/ui/useCanChange';
import styles from '../activity/DataSourcesArea.module.css';
import { GenerateCodeDialog } from './GenerateCodeDialog';
import { useDeactivateMediaCode, useDeleteMediaCode, useMediaCodes } from './api';
import { taggedLink } from './tagged-link';
import type { MediaCode } from './api';

/** The activity's Media codes section (FR-040): absent wherever the activity's type does not use codes. */
export function MediaCodesSection({ activityId }: { activityId: string }) {
  const area = useMediaCodes(activityId);
  if (area.data && !area.data.requiresCodes) {
    return null;
  }
  return (
    <Card title="Media codes">
      {area.isError ? (
        <ErrorState error={area.error} onRetry={() => void area.refetch()} />
      ) : area.data ? (
        <CodesBody activityId={activityId} items={area.data.items} />
      ) : (
        <Skeleton lines={3} />
      )}
    </Card>
  );
}

function CodesBody({ activityId, items }: { activityId: string; items: MediaCode[] }) {
  const canLink = useCanChange('performance.link');
  const [adding, setAdding] = useState(false);
  return (
    <div className={styles.stack}>
      {items.length === 0 ? (
        <p className={styles.notice}>
          No codes yet. Generate one to tag this activity&apos;s links.
        </p>
      ) : (
        <CodeRows activityId={activityId} items={items} canLink={canLink} />
      )}
      {canLink && (
        <div>
          <Button variant="secondary" onClick={() => setAdding(true)}>
            Generate code
          </Button>
        </div>
      )}
      {adding && <GenerateCodeDialog activityId={activityId} onClose={() => setAdding(false)} />}
    </div>
  );
}

function CodeRows({
  activityId,
  items,
  canLink,
}: {
  activityId: string;
  items: MediaCode[];
  canLink: boolean;
}) {
  const { show } = useToast();
  const deleteCode = useDeleteMediaCode(activityId);
  const deactivate = useDeactivateMediaCode(activityId);
  const [removing, setRemoving] = useState<MediaCode | null>(null);

  async function copy(text: string, what: string) {
    await navigator.clipboard.writeText(text);
    show({ message: `${what} copied` });
  }

  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th scope="col">Code</th>
          <th scope="col">Landing page</th>
          <th scope="col">Status</th>
          <th scope="col">
            <span className="visually-hidden">Actions</span>
          </th>
        </tr>
      </thead>
      <tbody>
        {items.map((code) => (
          <tr key={code.id}>
            <td className={styles.numeric}>{code.code}</td>
            <td>{code.landingPage || '—'}</td>
            <td>
              <StatusChip tone={code.status === 'active' ? 'positive' : 'neutral'}>
                {code.status}
              </StatusChip>
            </td>
            <td>
              <Button variant="tertiary" onClick={() => void copy(String(code.code), 'Code')}>
                Copy code
              </Button>
              <Button variant="tertiary" onClick={() => void copy(taggedLink(code), 'Tagged link')}>
                Copy tagged link
              </Button>
              {canLink && (
                <Button variant="tertiary" onClick={() => setRemoving(code)}>
                  Delete
                </Button>
              )}
            </td>
          </tr>
        ))}
        {removing && (
          <ConfirmDialog
            open
            onOpenChange={(open) => {
              if (!open) setRemoving(null);
            }}
            title="Delete this media code?"
            confirmLabel="Delete code"
            tone="danger"
            busy={deleteCode.isPending}
            onConfirm={() => {
              deleteCode.mutate(removing.id, {
                onSuccess: () => {
                  show({ message: `Code ${removing.code} deleted` });
                  setRemoving(null);
                },
                onError: (error) => {
                  show({ message: asApiError(error).message });
                  deactivate.mutate(removing.id, {
                    onSuccess: () => {
                      show({
                        message: `Code ${removing.code} deactivated — matching stops, its history stays`,
                      });
                      setRemoving(null);
                    },
                  });
                },
              });
            }}
          >
            Its number {removing.code} is never reused. A code already used by facts is deactivated
            instead.
          </ConfirmDialog>
        )}
      </tbody>
    </table>
  );
}
