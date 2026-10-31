import { ChartColumnIncreasing } from 'lucide-react';
import { ComingSoonPage } from '@/features/placeholders';
import { REPORT_ROUTES } from '../reports';
import type { ReportKey } from '../reports';

/** A report whose feature is not built yet; its lane replaces the stub that renders this. */
export function ComingSoonReport({ reportKey }: { reportKey: ReportKey }) {
  const report = REPORT_ROUTES.find((route) => route.key === reportKey);
  const name = report?.name ?? 'Report';
  return (
    <ComingSoonPage
      section={{
        label: name,
        icon: ChartColumnIncreasing,
        comingSoon: `This report arrives with feature ${report?.feature ?? ''} of the Reports area.`,
      }}
    />
  );
}
