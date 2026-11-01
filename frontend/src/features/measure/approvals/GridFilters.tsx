import { useEffect, useState } from 'react';
import { SearchField } from '@/ui/SearchField/SearchField';
import type { ApprovalsGridData } from './api';
import type { ApprovalsParams } from './schemas';
import { OUTCOME_WORDS, recordWord } from './words';
import styles from './ApprovalsGrid.module.css';

export interface GridFiltersProps {
  choices: ApprovalsGridData['choices'];
  params: ApprovalsParams;
  onChange: (changes: Partial<ApprovalsParams>) => void;
}

const SEARCH_DELAY_MS = 300;

function Choice(props: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (value: string) => void;
}) {
  return (
    <label className={styles.field}>
      {props.label}
      <select
        className={styles.select}
        value={props.value}
        onChange={(event) => props.onChange(event.target.value)}
      >
        <option value="">All</option>
        {props.options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

export function GridFilters({ choices, params, onChange }: GridFiltersProps) {
  const [typed, setTyped] = useState<string | null>(null);
  const text = typed ?? params.q;
  useEffect(() => {
    if (typed === null) {
      return;
    }
    const timer = setTimeout(() => {
      onChange({ q: typed.trim() });
      setTyped(null);
    }, SEARCH_DELAY_MS);
    return () => clearTimeout(timer);
  }, [typed, onChange]);
  return (
    <div className={styles.toolbar}>
      <SearchField
        label="Search approvals"
        placeholder="Search name, approver, owner or comment"
        value={text}
        onChange={setTyped}
      />
      <Choice
        label="Record type"
        value={params.recordTypes[0] ?? ''}
        options={choices.recordTypes.map((type) => ({ value: type, label: recordWord(type) }))}
        onChange={(value) => onChange({ recordTypes: value ? [value] : [] })}
      />
      <Choice
        label="Status"
        value={params.outcomes[0] ?? ''}
        options={choices.outcomes.map((o) => ({ value: o, label: OUTCOME_WORDS[o] ?? o }))}
        onChange={(value) => onChange({ outcomes: value ? [value] : [] })}
      />
      <Choice
        label="Approver"
        value={params.approvers[0] ?? ''}
        options={choices.approvers.map((p) => ({ value: p.id, label: p.name }))}
        onChange={(value) => onChange({ approvers: value ? [value] : [] })}
      />
    </div>
  );
}
