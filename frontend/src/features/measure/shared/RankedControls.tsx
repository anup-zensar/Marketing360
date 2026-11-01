import { useId } from 'react';
import { Button } from '@/ui/Button/Button';
import { SearchField } from '@/ui/SearchField/SearchField';
import { SegmentedFilter } from '@/ui/SegmentedFilter/SegmentedFilter';
import type { RankedState } from './ranked-params';
import styles from './RankedList.module.css';

export interface Choice {
  value: string;
  label: string;
}

interface RankedControlsProps {
  state: RankedState;
  sorts: Choice[];
  groups: Choice[];
  onChange: (patch: Partial<RankedState>) => void;
}

const ORDERS: { value: RankedState['order']; label: string }[] = [
  { value: 'desc', label: 'High to low' },
  { value: 'asc', label: 'Low to high' },
];

/** Sort by, Order, Group by, page size and the name search; each applies at once (040 FR-009, FR-013, FR-014). */
export function RankedControls({ state, sorts, groups, onChange }: RankedControlsProps) {
  const id = useId();
  return (
    <div className={styles.controls}>
      <label className={styles.field}>
        Sort by
        <select value={state.sort} onChange={(e) => onChange({ sort: e.target.value })}>
          {sorts.map((s) => (
            <option key={s.value} value={s.value}>
              {s.label}
            </option>
          ))}
        </select>
      </label>
      <SegmentedFilter
        label="Order"
        options={ORDERS}
        value={state.order}
        onChange={(order) => onChange({ order })}
      />
      <label className={styles.field}>
        Group by
        <select value={state.group} onChange={(e) => onChange({ group: e.target.value })}>
          {groups.map((g) => (
            <option key={g.value} value={g.value}>
              {g.label}
            </option>
          ))}
        </select>
      </label>
      <label className={styles.field} htmlFor={`${id}-size`}>
        Per page
        <select
          id={`${id}-size`}
          value={state.size}
          onChange={(e) =>
            onChange({ size: e.target.value === '50' ? 50 : e.target.value === '25' ? 25 : 10 })
          }
        >
          {[10, 25, 50].map((size) => (
            <option key={size} value={size}>
              {size}
            </option>
          ))}
        </select>
      </label>
      <SearchField label="Search by name" value={state.q} onChange={(q) => onChange({ q })} />
    </div>
  );
}

interface PagingProps {
  state: RankedState;
  total: number;
  onChange: (patch: Partial<RankedState>) => void;
}

export function Paging({ state, total, onChange }: PagingProps) {
  const pages = Math.max(1, Math.ceil(total / state.size));
  if (pages === 1) {
    return null;
  }
  return (
    <nav className={styles.paging} aria-label="Pages">
      <Button
        size="sm"
        disabled={state.page <= 1}
        onClick={() => onChange({ page: state.page - 1 })}
      >
        Previous
      </Button>
      <span>
        Page {state.page} of {pages}
      </span>
      <Button
        size="sm"
        disabled={state.page >= pages}
        onClick={() => onChange({ page: state.page + 1 })}
      >
        Next
      </Button>
    </nav>
  );
}
