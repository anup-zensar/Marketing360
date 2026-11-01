import { ChartColumnIncreasing } from 'lucide-react';
import { ComingSoonPage } from '@/features/placeholders';

export function ReportsPage() {
  return (
    <ComingSoonPage
      section={{
        label: 'Reports',
        icon: ChartColumnIncreasing,
        comingSoon:
          'Read-only reports of your plans, approvals, purchase requests and their figures.',
      }}
    />
  );
}
