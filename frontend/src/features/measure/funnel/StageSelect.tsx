import styles from './Funnel.module.css';

const STAGE_WORDS: Record<string, string> = { inq: 'INQ', mql: 'MQL' };

/** The funnel stage selector: every stage any model in view allows, in funnel order (049-C11, FR-015). */
export function StageSelect({
  stage,
  available,
  onChange,
}: {
  stage: string;
  available: string[];
  onChange: (stage: string) => void;
}) {
  return (
    <label className={styles.field}>
      Funnel stage
      <select value={stage} onChange={(e) => onChange(e.target.value)}>
        {available.map((s) => (
          <option key={s} value={s}>
            {STAGE_WORDS[s] ?? s}
          </option>
        ))}
      </select>
    </label>
  );
}
