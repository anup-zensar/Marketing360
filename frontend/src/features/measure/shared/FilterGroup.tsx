import { useId, useState } from 'react';
import { Button } from '@/ui/Button/Button';
import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { SearchField } from '@/ui/SearchField/SearchField';
import type { Ticked } from './filter-keys';
import styles from './ReportFrame.module.css';

export interface GroupOption {
  id: string;
  label: string;
}

interface FilterGroupProps {
  title: string;
  options: GroupOption[];
  ticked: Ticked;
  onChange: (ticked: Ticked) => void;
}

const SEARCH_ABOVE = 10;

/** One filter group: "(selected / all)", Select all, Clear all and, above ten values, a search (FR-010). */
export function FilterGroup({ title, options, ticked, onChange }: FilterGroupProps) {
  const [search, setSearch] = useState('');
  const id = useId();
  const isTicked = (value: string) => ticked === null || ticked.includes(value);
  const count = ticked === null ? options.length : ticked.length;
  const shown = options.filter((o) => o.label.toLowerCase().includes(search.trim().toLowerCase()));

  function toggle(value: string, checked: boolean) {
    const current = ticked ?? options.map((o) => o.id);
    const next = checked ? [...current, value] : current.filter((v) => v !== value);
    onChange(next.length === options.length ? null : next);
  }

  return (
    <fieldset className={styles.group}>
      <legend className={styles.groupTitle}>
        {title}{' '}
        <span className={styles.counter}>
          ({count} / {options.length})
        </span>
      </legend>
      <div className={styles.groupActions}>
        <Button type="button" size="sm" variant="tertiary" onClick={() => onChange(null)}>
          Select all
        </Button>
        <Button type="button" size="sm" variant="tertiary" onClick={() => onChange([])}>
          Clear all
        </Button>
      </div>
      {options.length > SEARCH_ABOVE && (
        <SearchField label={`Search ${title}`} value={search} onChange={setSearch} />
      )}
      <ul className={styles.values}>
        {shown.map((option) => (
          <li key={option.id}>
            <CheckboxField
              label={option.label}
              name={`${id}-${option.id}`}
              checked={isTicked(option.id)}
              onChange={(checked) => toggle(option.id, checked)}
            />
          </li>
        ))}
      </ul>
    </fieldset>
  );
}
