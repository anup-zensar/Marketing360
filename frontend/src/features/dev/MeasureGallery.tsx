import { ReportTable, ReportWidget } from '@/features/measure';
import type { ReportTableData } from '@/features/measure';
import { Card } from '@/ui/Card/Card';
import styles from './MeasureGallery.module.css';

const BY_BRAND: ReportTableData = {
  columns: [{ label: 'Brand' }, { label: 'Activities', numeric: true }],
  rows: [
    { key: 'helix', cells: ['Helix', '12'] },
    { key: 'pennzoil', cells: ['Pennzoil', '7'] },
    { key: 'none', cells: ['Not filled in', '3'] },
  ],
  total: ['Total', '22'],
};

const MOST = 12;

function BrandBars() {
  return (
    <svg className={styles.chart} viewBox="0 0 240 96" role="img" aria-label="Activities by brand">
      {BY_BRAND.rows.map((row, i) => {
        const value = Number(row.cells[1]);
        return (
          <g key={row.key} transform={`translate(0 ${i * 32})`}>
            <text x="0" y="18" className={styles.label}>
              {row.cells[0]}
            </text>
            <rect x="96" y="6" width={(value / MOST) * 128} height="16" className={styles.bar} />
          </g>
        );
      })}
    </svg>
  );
}

/** The Reports area's shared pieces (039 contracts/ui.md): a widget with its Chart · Table switch, and the table. */
export function MeasureGallery() {
  return (
    <Card title="ReportWidget and ReportTable">
      <ReportWidget
        report="planDetails"
        widget="galleryBrands"
        title="Activities by brand"
        tip="Activities per Brand value, Not filled in last."
        chart={<BrandBars />}
        table={BY_BRAND}
      />
      <ReportTable title="Activities by brand, as a table" data={BY_BRAND} />
    </Card>
  );
}
