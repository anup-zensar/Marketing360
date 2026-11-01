import { useSearchParams } from 'react-router';
import { ReportFrame } from '../shared';
import type { ReportAddress } from '../shared';
import { useApprovalsReport } from './api';
import { AverageWidget } from './AverageWidget';
import { readApprovalsParams, writeApprovalsParams } from './schemas';
import { StatusWidget } from './StatusWidget';

function Widgets({ address }: { address: ReportAddress }) {
  const [search, setSearch] = useSearchParams();
  const params = readApprovalsParams(search);
  const report = useApprovalsReport(address, params);
  const retry = () => void report.refetch();
  return (
    <>
      <StatusWidget
        data={report.data?.statusMix}
        statusType={params.statusType}
        onStatusType={(statusType) => setSearch(writeApprovalsParams(search, { statusType }))}
        failed={report.isError}
        onRetry={retry}
      />
      <AverageWidget data={report.data?.average} failed={report.isError} onRetry={retry} />
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
