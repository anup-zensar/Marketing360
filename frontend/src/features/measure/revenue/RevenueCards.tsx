import { useFormatters } from '@/format/useFormatters';
import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import { reasonText } from './reasons';
import type { RevenueCards as RevenueCardsData } from './api';
import styles from './Revenue.module.css';

/** The fixed Revenue card set: goal, actual, projected, performance (049-C9). */
export function RevenueCards({ cards }: { cards: RevenueCardsData }) {
  const { money } = useFormatters();
  return (
    <div className={styles.cards} aria-label="Summary">
      <SummaryCard
        label="Revenue goal"
        value={cards.goal.value ? money(cards.goal.value) : reasonText(cards.goal.reason)}
      />
      <SummaryCard
        label="Actual revenue"
        value={cards.actual.value ? money(cards.actual.value) : reasonText(cards.actual.reason)}
        tip="Revenue facts matched to the activities in view, to today."
      />
      <SummaryCard
        label="Actual or projected"
        value={
          cards.projected.value ? money(cards.projected.value) : reasonText(cards.projected.reason)
        }
        tip="Actual to the end of last period, projected from today."
      />
      <SummaryCard
        label="Performance"
        value={
          cards.performance.value !== null
            ? `${cards.performance.value}%`
            : reasonText(cards.performance.reason)
        }
        tip="Actual or projected ÷ goal."
      />
    </div>
  );
}
