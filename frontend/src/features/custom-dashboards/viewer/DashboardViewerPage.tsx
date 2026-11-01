import { Copy, LayoutDashboard, Pencil, RotateCw } from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router';
import { ApiError } from '@/api/errors';
import { isShown, useAccess } from '@/session/access';
import { Button, LinkButton } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { TeamPills } from '@/ui/TeamPill/TeamPill';
import { usePageTitle } from '@/ui/usePageTitle';
import type { CustomDashboard } from '../model';
import { useCustomDashboard, useRefreshWidgets } from './api';
import { DeleteDashboardDialog } from './DeleteDashboardDialog';
import { FiltersBar } from './FiltersBar';
import { hasAddressFilters, readFilters, writeFilters } from './filters-address';
import { useCopyDashboard } from './useCopyDashboard';
import { ViewerGrid } from './ViewerGrid';
import styles from './DashboardViewerPage.module.css';

const LIST_CRUMB = { label: 'Dashboards', to: '/dashboards' };

export function DashboardViewerPage() {
  const { dashboardId = '' } = useParams();
  const query = useCustomDashboard(dashboardId);
  usePageTitle(query.data?.name ?? 'Dashboard');
  const failure = query.error instanceof ApiError ? query.error : null;

  if (query.data) {
    return <Viewer dashboard={query.data} />;
  }
  if (failure?.kind === 'forbidden') {
    return <NoAccess message={failure.message} />;
  }
  return (
    <>
      <PageHeader breadcrumb={[LIST_CRUMB, { label: 'Dashboard' }]} title="Dashboard" />
      <Card>
        {failure?.kind === 'notFound' ? (
          <EmptyState
            icon={<LayoutDashboard />}
            title="Dashboard not found"
            text="It may have been deleted, or it isn't shared with you."
            action={<LinkButton to="/dashboards">Go to Dashboards</LinkButton>}
          />
        ) : query.isError ? (
          <ErrorState error={query.error} onRetry={() => void query.refetch()} />
        ) : (
          <div aria-label="Loading the dashboard">
            <Skeleton lines={6} height="24px" />
          </div>
        )}
      </Card>
    </>
  );
}

function Viewer({ dashboard }: { dashboard: CustomDashboard }) {
  const [params, setParams] = useSearchParams();
  const navigate = useNavigate();
  const { has } = useAccess();
  const refresh = useRefreshWidgets();
  const { copyAndOpen, isCopying } = useCopyDashboard();
  const [isDeleting, setDeleting] = useState(false);
  const filters = readFilters(params, dashboard.filters);
  const { access } = dashboard;

  return (
    <>
      <PageHeader
        breadcrumb={[LIST_CRUMB, { label: dashboard.name }]}
        title={dashboard.name}
        meta={<Facts dashboard={dashboard} />}
        description={dashboard.description || undefined}
        moreActions={
          isShown(access.delete)
            ? [{ label: 'Delete', tone: 'danger', onSelect: () => setDeleting(true) }]
            : undefined
        }
        actions={
          <>
            <Button icon={<RotateCw />} onClick={() => void refresh()}>
              Refresh
            </Button>
            {has('dashboards.build') && isShown(access.copy) && (
              <Button
                icon={<Copy />}
                aria-busy={isCopying || undefined}
                onClick={() => void copyAndOpen(dashboard.id)}
              >
                Make a copy
              </Button>
            )}
            {dashboard.isOwner && isShown(access.edit) && (
              <LinkButton
                to={`/builder/${encodeURIComponent(dashboard.id)}`}
                variant="primary"
                icon={<Pencil />}
              >
                Edit
              </LinkButton>
            )}
          </>
        }
      />
      <div className={styles.page}>
        <FiltersBar
          filters={filters}
          canReset={hasAddressFilters(params)}
          onChange={(next) => setParams(writeFilters(next))}
          onReset={() => setParams({})}
        />
        {dashboard.widgets.length === 0 ? (
          <Card>
            <EmptyState
              icon={<LayoutDashboard />}
              title="This dashboard has no widgets yet"
              text={
                dashboard.isOwner
                  ? 'Choose Edit to add charts, figures and notes.'
                  : 'Its owner has not added any charts yet.'
              }
            />
          </Card>
        ) : (
          <ViewerGrid
            widgets={dashboard.widgets}
            filters={filters}
            dashboardName={dashboard.name}
            viewerOwnsDashboard={dashboard.isOwner}
          />
        )}
      </div>
      {isDeleting && (
        <DeleteDashboardDialog
          dashboard={dashboard}
          onClose={() => setDeleting(false)}
          onDeleted={() => void navigate('/dashboards')}
        />
      )}
    </>
  );
}

function Facts({ dashboard }: { dashboard: CustomDashboard }) {
  return (
    <div className={styles.facts}>
      <span>By {dashboard.owner.name}</span>
      {dashboard.sharedWith.length > 0 ? (
        <TeamPills teams={dashboard.sharedWith} />
      ) : (
        <span>Not shared</span>
      )}
    </div>
  );
}
