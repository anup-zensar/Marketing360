import { useActivityCategories } from '@/features/settings';
import styles from './Revenue.module.css';

const ALL = '';
const UNCATEGORISED = 'uncategorised';

/** The activity category control (049-C10): narrows the cards, the series and the ranked list. */
export function CategorySelect({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  const categories = useActivityCategories();
  return (
    <label className={styles.field}>
      Activity category
      <select value={value} onChange={(e) => onChange(e.target.value)}>
        <option value={ALL}>All categories</option>
        {categories.data?.items.map((c) => (
          <option key={c.id} value={c.id}>
            {c.name}
          </option>
        ))}
        <option value={UNCATEGORISED}>Uncategorised</option>
      </select>
    </label>
  );
}
