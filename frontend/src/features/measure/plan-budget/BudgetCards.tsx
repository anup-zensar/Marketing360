import { useFormatters } from '@/format/useFormatters';
import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import type { PlanBudgetCards } from './api';
import styles from './PlanBudget.module.css';

const NONE = '—';

/** The ten cards of 040 FR-002; "—" where a figure has no value or the person may not see it. */
export function BudgetCards({ cards }: { cards: PlanBudgetCards }) {
  const { money } = useFormatters();
  const amount = (value: PlanBudgetCards['budget']) => (value ? money(value) : NONE);
  const percent = (value: number | null) => (value === null ? NONE : `${value}%`);
  const overAllocated = cards.leftToPlan !== null && cards.leftToPlan.amount.startsWith('-');
  return (
    <div className={styles.cards} aria-label="Summary">
      <SummaryCard label="Budget" value={amount(cards.budget)} tip="Counted once per branch." />
      <SummaryCard label="Planned cost" value={amount(cards.planned)} />
      <SummaryCard
        label="Committed"
        value={amount(cards.committed)}
        tip="Purchase-order lines of the year."
      />
      <SummaryCard label="Actual" value={amount(cards.actual)} tip="Payments posted in the year." />
      <SummaryCard
        label="Budget used"
        value={percent(cards.budgetUsed)}
        tip="Planned cost ÷ budget."
      />
      <SummaryCard
        label="Left to plan"
        value={amount(cards.leftToPlan)}
        tone={overAllocated ? 'problem' : undefined}
        detail={overAllocated ? 'Over-allocated' : undefined}
      />
      <SummaryCard label="Unspent budget" value={amount(cards.unspent)} tip="Budget − actual." />
      <SummaryCard label="Budget spent" value={percent(cards.budgetSpent)} tip="Actual ÷ budget." />
      <SummaryCard label="Not yet committed" value={amount(cards.notYetCommitted)} />
      <SummaryCard
        label="% committed"
        value={percent(cards.committedShare)}
        tip="Committed ÷ planned cost."
      />
    </div>
  );
}
