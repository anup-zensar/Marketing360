import styles from './Inspector.module.css';

export interface NumberInputProps {
  label: string;
  name: string;
  value: number;
  min: number;
  max: number;
  error?: string | undefined;
  onChange: (value: number) => void;
}

export function NumberInput({ label, name, value, min, max, error, onChange }: NumberInputProps) {
  return (
    <div className={styles.number}>
      <label htmlFor={name}>{label}</label>
      <input
        id={name}
        type="number"
        min={min}
        max={max}
        step={1}
        value={Number.isFinite(value) ? value : ''}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${name}-error` : undefined}
        onChange={(event) => onChange(event.target.valueAsNumber)}
      />
      {error && (
        <p id={`${name}-error`} className={styles.error}>
          {error}
        </p>
      )}
    </div>
  );
}
