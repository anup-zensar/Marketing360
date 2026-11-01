import { useSearchParams } from 'react-router';
import { ReportFrame } from '../shared';
import type { ReportAddress } from '../shared';
import { useApprovalsReport } from './api';
import { ApprovalsGrid } from './ApprovalsGrid';
import { AverageWidget } from './AverageWidget';
import { readApprovalsParams, writeApprovalsParams } from './schemas';
import type { ApprovalsParams } from './schemas';
import { StatusWidget } from './StatusWidget';

function Widgets({ address }: { address: ReportAddress }) {
  const [search, setSearch] = useSearchParams();
  const params = readApprovalsParams(search);
  const report = useApprovalsReport(address, params);
  const retry = () => void report.refetch();
  const change = (changes: Partial<ApprovalsParams>) =>
    setSearch(writeApprovalsParams(search, { ...params, page: 1, ...changes }));
  return (
    <>
      <StatusWidget
        data={report.data?.statusMix}
        statusType={params.statusType}
        onStatusType={(statusType) => change({ statusType, page: params.page })}
        failed={report.isError}
        onRetry={retry}
      />
      <AverageWidget data={report.data?.average} failed={report.isError} onRetry={retry} />
      <ApprovalsGrid
        grid={report.data?.grid}
        covers={report.data?.financeCovers}
        params={params}
        onParams={change}
        failed={report.isError}
        onRetry={retry}
      />
    </>
  );
}

export function ApprovalsReport() {
  return (
    <ReportFrame report="approvals">
      {(_frame, address) => <Widgets address={address} />}
    </ReportFrame>
  );
}
