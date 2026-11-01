import { REPORT_ROUTES } from '../reports';
import type { ReportKey } from '../reports';
import { ReportFrame } from './ReportFrame';

/** A report whose feature is not built yet: the frame, with "Coming soon" in place of its widgets (039 T030). */
export function ComingSoonReport({ reportKey }: { reportKey: ReportKey }) {
  const report = REPORT_ROUTES.find((route) => route.key === reportKey);
  return (
    <ReportFrame
      report={reportKey}
      comingSoon={`Coming soon — this report arrives with feature ${report?.feature ?? ''}.`}
    />
  );
}
