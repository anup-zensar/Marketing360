import { Pencil, Share2, Table2 } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router';
import { ApiError, isSessionLost } from '@/api/errors';
import { plural } from '@/format/text';
import { isShown } from '@/session/access';
import { Button, LinkButton } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { TeamPills } from '@/ui/TeamPill/TeamPill';
import { useToast } from '@/ui/Toast/useToast';
import { usePageTitle } from '@/ui/usePageTitle';
import type { Dataset } from '../model';
import { SharingDialog } from '../sharing/SharingDialog';
import { useDataset, useShareDataset } from './api';
import { DatasetDetailsForm } from './DatasetDetailsForm';
import { DeleteDatasetDialog } from './DeleteDatasetDialog';
import { RowsTable } from './RowsTable';
import styles from './datasets.module.css';

const CRUMBS = [
  { label: 'Dashboards', to: '/dashboards' },
  { label: 'Datasets', to: '/datasets' },
];
const KIND_LABELS = { text: 'Text', number: 'Number', date: 'Date', money: 'Money' } as const;

export function DatasetPage() {
  const { datasetId = '' } = useParams();
  const query = useDataset(datasetId);
  usePageTitle(query.data?.name ?? 'Dataset');
  const failure = query.error instanceof ApiError ? query.error : null;

  if (query.data) {
    return <DatasetDetails dataset={query.data} />;
  }
  return (
    <>
      <PageHeader breadcrumb={[...CRUMBS, { label: 'Dataset' }]} title="Dataset" />
      <Card>
        {failure?.kind === 'notFound' ? (
          <EmptyState
            icon={<Table2 />}
            title="Dataset not found"
            text="It may have been deleted, or it isn't shared with you."
            action={<LinkButton to="/datasets">Go to Datasets</LinkButton>}
          />
        ) : query.isError ? (
          <ErrorState error={query.error} onRetry={() => void query.refetch()} />
        ) : (
          <div aria-label="Loading the dataset">
            <Skeleton lines={6} height="20px" />
          </div>
        )}
      </Card>
    </>
  );
}

type Open = 'edit' | 'share' | 'delete' | null;

function DatasetDetails({ dataset }: { dataset: Dataset }) {
  const navigate = useNavigate();
  const { show } = useToast();
  const share = useShareDataset();
  const [open, setOpen] = useState<Open>(null);
  const [shareError, setShareError] = useState<string | null>(null);
  const { access } = dataset;

  async function saveSharing(teamIds: string[]) {
    setShareError(null);
    try {
      await share.mutateAsync({ id: dataset.id, version: dataset.version, teamIds });
      show({ message: 'Sharing saved' });
      setOpen(null);
    } catch (error) {
      const failure = error instanceof ApiError ? error : new ApiError('other');
      if (!isSessionLost(failure)) {
        setShareError(failure.message);
      }
    }
  }

  return (
    <>
      <PageHeader
        breadcrumb={[...CRUMBS, { label: dataset.name }]}
        title={dataset.name}
        meta={<Facts dataset={dataset} />}
        description={dataset.description || undefined}
        moreActions={
          isShown(access.delete)
            ? [{ label: 'Delete', tone: 'danger', onSelect: () => setOpen('delete') }]
            : undefined
        }
        actions={
          <>
            {isShown(access.share) && (
              <Button icon={<Share2 />} onClick={() => setOpen('share')}>
                Sharing
              </Button>
            )}
            {isShown(access.edit) && (
              <Button variant="primary" icon={<Pencil />} onClick={() => setOpen('edit')}>
                Edit
              </Button>
            )}
          </>
        }
      />
      <div className={styles.sections}>
        <Card title="Columns">
          <ColumnsTable dataset={dataset} />
        </Card>
        <Card title="Rows">
          <RowsTable datasetId={dataset.id} />
        </Card>
        <Card title="Used by">
          <UsedBy dataset={dataset} />
        </Card>
      </div>
      {open === 'edit' && <DatasetDetailsForm dataset={dataset} onClose={() => setOpen(null)} />}
      <SharingDialog
        open={open === 'share'}
        recordLabel="dataset"
        recordName={dataset.name}
        teams={dataset.shareableTeams}
        selected={dataset.sharedWith.map((team) => team.id)}
        saving={share.isPending}
        error={shareError}
        onSave={(teamIds) => void saveSharing(teamIds)}
        onClose={() => {
          setShareError(null);
          setOpen(null);
        }}
      />
      {open === 'delete' && (
        <DeleteDatasetDialog
          dataset={dataset}
          onClose={() => setOpen(null)}
          onDeleted={() => void navigate('/datasets')}
        />
      )}
    </>
  );
}

function Facts({ dataset }: { dataset: Dataset }) {
  const source = dataset.sheetName
    ? `${dataset.fileName}, sheet '${dataset.sheetName}'`
    : dataset.fileName;
  return (
    <div className={styles.facts}>
      <p>
        {`By ${dataset.owner.name} · Uploaded from ${source} · ${plural(dataset.rowCount, 'row')}`}
      </p>
      {dataset.sharedWith.length > 0 ? <TeamPills teams={dataset.sharedWith} /> : <p>Not shared</p>}
    </div>
  );
}

function ColumnsTable({ dataset }: { dataset: Dataset }) {
  return (
    <div role="region" aria-label="Columns" tabIndex={0} className={styles.scroller}>
      <table className={styles.table}>
        <caption className="visually-hidden">Columns of the dataset</caption>
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Type</th>
            <th scope="col">Currency</th>
          </tr>
        </thead>
        <tbody>
          {dataset.columns.map((column) => (
            <tr key={column.key}>
              <td>{column.label}</td>
              <td>{KIND_LABELS[column.kind]}</td>
              <td>{column.currency ?? '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function UsedBy({ dataset }: { dataset: Dataset }) {
  if (dataset.usedBy.length === 0) {
    return <p className={styles.facts}>No dashboard you can see uses this dataset.</p>;
  }
  return (
    <ul role="list" className={styles.usedBy}>
      {dataset.usedBy.map((dashboard) => (
        <li key={dashboard.id}>
          <Link to={`/dashboards/${encodeURIComponent(dashboard.id)}`}>{dashboard.name}</Link>
        </li>
      ))}
    </ul>
  );
}
