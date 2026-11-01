import { LayoutDashboard, Plus, Share2 } from 'lucide-react';
import { useId, useState } from 'react';
import type { ReactNode } from 'react';
import { useAccess } from '@/session/access';
import { LinkButton } from '@/ui/Button/Button';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { usePageTitle } from '@/ui/usePageTitle';
import type { CustomDashboardSummary } from '../model';
import { useCustomDashboards } from './api';
import { BuiltInCards, DashboardCard } from './DashboardCard';
import { DeleteDashboardDialog } from './DeleteDashboardDialog';
import { useCopyDashboard } from './useCopyDashboard';
import styles from './DashboardListPage.module.css';

export function DashboardListPage() {
  usePageTitle('Dashboards');
  const { has } = useAccess();
  const canSeePlans = has('plans.view');
  const canView = has('dashboards.view');
  const canBuild = has('dashboards.build');

  if (!canSeePlans && !canView) {
    return <NoAccess />;
  }

  return (
    <>
      <PageHeader
        title="Dashboards"
        description="The built-in dashboards and those you and your teams build."
        actions={
          canBuild && (
            <>
              <LinkButton to="/datasets">Datasets</LinkButton>
              <LinkButton to="/builder" variant="primary" icon={<Plus />}>
                New dashboard
              </LinkButton>
            </>
          )
        }
      />
      <div className={styles.sections}>
        {canSeePlans && (
          <Section title="Built-in">
            <BuiltInCards />
          </Section>
        )}
        {canView && <CustomSections canBuild={canBuild} />}
      </div>
    </>
  );
}

function CustomSections({ canBuild }: { canBuild: boolean }) {
  const list = useCustomDashboards();
  const { copyAndOpen } = useCopyDashboard();
  const [deleting, setDeleting] = useState<CustomDashboardSummary | null>(null);

  if (!list.data) {
    return (
      <Section title="My dashboards">
        {list.isError ? (
          <ErrorState error={list.error} onRetry={() => void list.refetch()} />
        ) : (
          <div aria-label="Loading dashboards" className={styles.cards}>
            <Skeleton height="120px" />
            <Skeleton height="120px" />
          </div>
        )}
      </Section>
    );
  }

  const cards = (items: CustomDashboardSummary[]) => (
    <ul role="list" className={styles.cards}>
      {items.map((dashboard) => (
        <DashboardCard
          key={dashboard.id}
          dashboard={dashboard}
          canBuild={canBuild}
          onCopy={(id) => void copyAndOpen(id)}
          onDelete={setDeleting}
        />
      ))}
    </ul>
  );

  return (
    <>
      <Section title="My dashboards">
        {list.data.mine.length > 0 ? (
          cards(list.data.mine)
        ) : (
          <EmptyState
            icon={<LayoutDashboard />}
            title="No dashboards yet"
            text="Build one from your plans or your own data."
            action={canBuild && <LinkButton to="/builder">New dashboard</LinkButton>}
          />
        )}
      </Section>
      <Section title="Shared with me">
        {list.data.sharedWithMe.length > 0 ? (
          cards(list.data.sharedWithMe)
        ) : (
          <EmptyState icon={<Share2 />} title="Nothing shared with you yet." />
        )}
      </Section>
      {deleting && <DeleteDashboardDialog dashboard={deleting} onClose={() => setDeleting(null)} />}
    </>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  const titleId = useId();
  return (
    <section className={styles.section} aria-labelledby={titleId}>
      <h2 id={titleId} className={styles.sectionTitle}>
        {title}
      </h2>
      {children}
    </section>
  );
}
