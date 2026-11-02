import styles from './RoiValue.module.css';

export interface RoiValueProps {
  /** Whole per cent from the API; null when there is no cost to divide by. */
  roi: number | null;
  model: string;
  basis: 'actual' | 'planned';
}

/** An ROI with what it rests on (050 FR-008): the model that credited the revenue and the cost it divides by. */
export function RoiValue({ roi, model, basis }: RoiValueProps) {
  const cost = basis === 'planned' ? 'on planned cost' : 'on actual cost';
  return (
    <span className={styles.roi}>
      <span className={styles.value}>{roi === null ? '—' : `${roi}%`}</span>
      <span className={styles.basis}>
        {roi === null && basis === 'actual'
          ? `No actual cost yet · ${model}`
          : `${model} · ${cost}`}
      </span>
    </span>
  );
}
