import { Share2, Table2, Upload } from 'lucide-react';
import { useState } from 'react';
import { Link, Navigate } from 'react-router';
import { formatWhole } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { usePageTitle } from '@/ui/usePageTitle';
import type { DatasetSummary } from '../model';
import { useDatasets } from './api';
import { UploadDialog } from './UploadDialog';
import styles from './datasets.module.css';

export function DatasetsPage() {
  usePageTitle('Datasets');
  const list = useDatasets();
  const [isUploading, setUploading] = useState(false);
  // The new dataset opens once the upload panel and its unsaved-changes guard are gone.
  const [savedId, setSavedId] = useState<string | null>(null);
  const upload = () => setUploading(true);

  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Dashboards', to: '/dashboards' }, { label: 'Datasets' }]}
        title="Datasets"
        description="Data from your own CSV and Excel files, to chart beside SP&A One data."
        actions={
          <Button variant="primary" icon={<Upload />} onClick={upload}>
            Upload data
          </Button>
        }
      />
      {list.data ? (
        <div className={styles.sections}>
          <Card title="My datasets">
            {list.data.mine.length > 0 ? (
              <DatasetTable datasets={list.data.mine} caption="My datasets" />
            ) : (
              <EmptyState
                icon={<Table2 />}
                title="No datasets yet"
                text="Upload a CSV or Excel file to chart data SP&A One doesn't hold."
                action={<Button onClick={upload}>Upload data</Button>}
              />
            )}
          </Card>
          <Card title="Shared with me">
            {list.data.sharedWithMe.length > 0 ? (
              <DatasetTable datasets={list.data.sharedWithMe} caption="Shared with me" withOwner />
            ) : (
              <EmptyState icon={<Share2 />} title="Nothing shared with you yet." />
            )}
          </Card>
        </div>
      ) : (
        <Card>
          {list.isError ? (
            <ErrorState error={list.error} onRetry={() => void list.refetch()} />
          ) : (
            <div aria-label="Loading datasets">
              <Skeleton lines={4} height="20px" />
            </div>
          )}
        </Card>
      )}
      {isUploading && (
        <UploadDialog
          onClose={() => setUploading(false)}
          onSaved={(dataset) => {
            setSavedId(dataset.id);
            setUploading(false);
          }}
        />
      )}
      {savedId && !isUploading && <Navigate to={`/datasets/${encodeURIComponent(savedId)}`} />}
    </>
  );
}

interface DatasetTableProps {
  datasets: DatasetSummary[];
  caption: string;
  withOwner?: boolean;
}

function DatasetTable({ datasets, caption, withOwner = false }: DatasetTableProps) {
  const { dateTime } = useFormatters();
  return (
    <div role="region" aria-label={caption} tabIndex={0} className={styles.scroller}>
      <table className={styles.table}>
        <caption className="visually-hidden">{caption}</caption>
        <thead>
          <tr>
            <th scope="col">Name</th>
            {withOwner && <th scope="col">Owner</th>}
            <th scope="col">Rows</th>
            <th scope="col">Columns</th>
            <th scope="col">Shared with</th>
            <th scope="col">Updated</th>
          </tr>
        </thead>
        <tbody>
          {datasets.map((dataset) => (
            <tr key={dataset.id}>
              <td>
                <Link to={`/datasets/${encodeURIComponent(dataset.id)}`}>{dataset.name}</Link>
              </td>
              {withOwner && <td>{dataset.owner.name}</td>}
              <td>{formatWhole(dataset.rowCount)}</td>
              <td>{formatWhole(dataset.columnCount)}</td>
              <td>{dataset.sharedWith.map((team) => team.name).join(', ') || 'Not shared'}</td>
              <td>{dateTime(dataset.updatedAt)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
