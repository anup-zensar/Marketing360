import { SegmentedFilter } from '@/ui/SegmentedFilter/SegmentedFilter';
import type { ReportChoices } from './api';
import type { Period, Show } from './report-address';
import { SHOW_WORDS, SHOWS } from './report-text';
import styles from './ReportFrame.module.css';

export function InViewLine({ text }: { text: string }) {
  return <p className={styles.inView}>{text}</p>;
}

export function PeriodControl({
  value,
  onChange,
}: {
  value: Period;
  onChange: (period: Period) => void;
}) {
  return (
    <SegmentedFilter
      label="Period"
      options={[
        { value: 'month', label: 'Months' },
        { value: 'quarter', label: 'Quarters' },
      ]}
      value={value}
      onChange={onChange}
    />
  );
}

export function ShowControl({ value, onChange }: { value: Show; onChange: (show: Show) => void }) {
  return (
    <SegmentedFilter
      label="Show"
      options={SHOWS.map((show) => ({
        value: show,
        label: SHOW_WORDS[show],
      }))}
      value={value}
      onChange={onChange}
    />
  );
}

interface CurrencyControlProps {
  currencies: ReportChoices['currencies'];
  value: string;
  onChange: (code: string) => void;
}

export function CurrencyControl({ currencies, value, onChange }: CurrencyControlProps) {
  return (
    <label className={styles.field}>
      Currency
      <select className={styles.select} value={value} onChange={(e) => onChange(e.target.value)}>
        {currencies.map((c) => (
          <option key={c.code} value={c.code}>
            {c.code} – {c.name}
          </option>
        ))}
      </select>
    </label>
  );
}
