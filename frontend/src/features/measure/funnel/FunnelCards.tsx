import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import { reasonText } from './reasons';
import type { FunnelCards as FunnelCardsData } from '../revenue/api';
import styles from './Funnel.module.css';

const STAGE_WORDS: Record<string, string> = { inq: 'INQ', mql: 'MQL' };

/** The fixed Funnel card set for one stage: goal, actual, projected, performance, conversion (049-C9). */
export function FunnelCards({ stage, cards }: { stage: string; cards: FunnelCardsData }) {
  const name = STAGE_WORDS[stage] ?? stage;
  const whole = (value: number | null) =>
    value === null ? null : Math.round(value).toLocaleString();
  return (
    <div className={styles.cards} aria-label="Summary">
      <SummaryCard
        label={`${name} goal`}
        value={whole(cards.goal.value) ?? reasonText(cards.goal.reason)}
      />
      <SummaryCard
        label={`${name} actual`}
        value={whole(cards.actual.value) ?? reasonText(cards.actual.reason)}
        tip="Facts matched to the activities in view, to today."
      />
      <SummaryCard
        label={`${name} actual or projected`}
        value={whole(cards.projected.value) ?? reasonText(cards.projected.reason)}
      />
      <SummaryCard
        label="Performance"
        value={
          cards.performance.value !== null
            ? `${cards.performance.value}%`
            : reasonText(cards.performance.reason)
        }
      />
      <SummaryCard
        label="Conversion"
        value={
          cards.conversion.value !== null
            ? `${cards.conversion.value}%`
            : reasonText(cards.conversion.reason)
        }
        tip="This stage's actual over the previous stage's actual."
      />
    </div>
  );
}
