import { SearchX, UsersRound } from 'lucide-react';
import { formatWhole } from '@/format/decimal';
import { Button } from '@/ui/Button/Button';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import { ReportFrame } from '../shared';
import { AccessControls } from './AccessControls';
import { AccessTable } from './AccessTable';
import { useAccessTeams, useUserAccess } from './api';
import type { UserAccessReport as Report } from './api';
import { useAccessTable } from './access-address';
import type { AccessTable as Table } from './access-address';
import { SignInBands } from './SignInBands';
import styles from './UserAccess.module.css';

/** User access (feature 045): everyone with roles, teams and last sign-in, in 039's page without plan filters. */
export function UserAccessReport() {
  return (
    <ReportFrame report="userAccess" filters={false}>
      {() => <UserAccessBody />}
    </ReportFrame>
  );
}

function UserAccessBody() {
  const { table, change } = useAccessTable();
  const report = useUserAccess(table);
  const teams = useAccessTeams();
  return (
    <div className={styles.body}>
      <AccessControls table={table} teams={teams.data ?? []} onChange={change} />
      {report.data ? (
        <Results report={report.data} table={table} onChange={change} />
      ) : report.isError ? (
        <ErrorState error={report.error} onRetry={() => void report.refetch()} />
      ) : (
        <div aria-label="Loading the people">
          <Skeleton lines={6} height="24px" />
        </div>
      )}
    </div>
  );
}

interface ResultsProps {
  report: Report;
  table: Table;
  onChange: (patch: Partial<Table>) => void;
}

/** The people, or why none match: the search, else the team and status filters (FR-012). */
function Results({ report, table, onChange }: ResultsProps) {
  if (report.total === 0 && table.q !== '') {
    return (
      <EmptyState
        icon={<SearchX />}
        title={`No people match "${table.q}"`}
        action={
          <Button type="button" variant="secondary" onClick={() => onChange({ q: '' })}>
            Clear search
          </Button>
        }
      />
    );
  }
  if (report.total === 0) {
    return (
      <EmptyState
        icon={<UsersRound />}
        title="No people match these filters"
        action={
          <Button
            type="button"
            variant="secondary"
            onClick={() => onChange({ teams: [], status: 'active' })}
          >
            Reset filters
          </Button>
        }
      />
    );
  }
  return (
    <>
      <div className={styles.cards} aria-label="Summary">
        <SummaryCard label="People" value={formatWhole(report.cards.people)} />
        <SummaryCard
          label="Signed in within 30 days"
          value={formatWhole(report.cards.signedInWithin30Days)}
        />
        <SummaryCard
          label="Never signed in"
          value={formatWhole(report.cards.neverSignedIn)}
          tip="Not since sign-in times were first recorded."
        />
      </div>
      <SignInBands
        bands={report.bands}
        people={report.cards.people}
        recordedSince={report.recordedSince}
      />
      <AccessTable rows={report.rows} total={report.total} table={table} onChange={onChange} />
    </>
  );
}
