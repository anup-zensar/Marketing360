import { useId } from 'react';
import { Button } from '@/ui/Button/Button';
import type { ApprovalsGridData } from './api';
import type { ApprovalsParams, PageSize } from './schemas';
import styles from './ApprovalsGrid.module.css';

export interface GridPagerProps {
  grid: ApprovalsGridData;
  sizes: readonly PageSize[];
  onParams: (changes: Partial<ApprovalsParams>) => void;
}

export function GridPager({ grid, sizes, onParams }: GridPagerProps) {
  const sizeId = useId();
  const from = grid.total === 0 ? 0 : (grid.page - 1) * grid.pageSize + 1;
  const to = Math.min(grid.page * grid.pageSize, grid.total);
  return (
    <nav className={styles.pager} aria-label="Pages">
      <p className={styles.range}>
        {grid.total === 0 ? '0 rows' : `Rows ${from}–${to} of ${grid.total}`}
      </p>
      <div className={styles.buttons}>
        <Button
          size="sm"
          disabled={grid.page <= 1}
          onClick={() => onParams({ page: grid.page - 1 })}
        >
          Previous
        </Button>
        <Button
          size="sm"
          disabled={to >= grid.total}
          onClick={() => onParams({ page: grid.page + 1 })}
        >
          Next
        </Button>
      </div>
      <label htmlFor={sizeId} className={styles.field}>
        Rows per page
        <select
          id={sizeId}
          className={styles.select}
          value={grid.pageSize}
          onChange={(event) => {
            const size = sizes.find((one) => one === Number(event.target.value));
            if (size !== undefined) {
              onParams({ pageSize: size });
            }
          }}
        >
          {sizes.map((size) => (
            <option key={size} value={size}>
              {size}
            </option>
          ))}
        </select>
      </label>
    </nav>
  );
}
