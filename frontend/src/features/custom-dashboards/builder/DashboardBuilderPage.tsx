import { useParams } from 'react-router';
import { LinkButton } from '@/ui/Button/Button';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { usePageTitle } from '@/ui/usePageTitle';
import { useDashboard } from './api';
import { BuilderEditor } from './BuilderEditor';
import { ViewOnlyBuilder } from './ViewOnlyBuilder';

export const NOT_FOUND_TITLE = 'Dashboard not found';

export function DashboardBuilderPage() {
  const { dashboardId } = useParams();
  const query = useDashboard(dashboardId);

  if (!dashboardId) {
    return <BuilderEditor key="new" dashboard={null} />;
  }
  if (query.isPending) {
    return <BuilderLoading />;
  }
  if (query.isError) {
    if (query.error.kind === 'notFound') {
      return <DashboardNotFound />;
    }
    return query.error.kind === 'forbidden' ? (
      <NoAccess />
    ) : (
      <>
        <PageHeader title="Edit dashboard" />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </>
    );
  }
  const dashboard = query.data;
  return dashboard.isOwner ? (
    <BuilderEditor key={dashboard.id} dashboard={dashboard} />
  ) : (
    <ViewOnlyBuilder dashboard={dashboard} />
  );
}

function BuilderLoading() {
  usePageTitle('Edit dashboard');
  return (
    <div aria-busy="true">
      <PageHeader title="Edit dashboard" />
      <p className="visually-hidden">Loading the dashboard…</p>
      <Skeleton height="480px" />
    </div>
  );
}

function DashboardNotFound() {
  usePageTitle(NOT_FOUND_TITLE);
  return (
    <PageHeader
      title={NOT_FOUND_TITLE}
      description="It may have been deleted, or it isn't shared with you."
      actions={
        <LinkButton to="/dashboards" variant="primary">
          Go to Dashboards
        </LinkButton>
      }
    />
  );
}
