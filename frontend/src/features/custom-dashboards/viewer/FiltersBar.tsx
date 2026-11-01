import { RotateCcw } from 'lucide-react';
import { useId } from 'react';
import { plural } from '@/format/text';
import { Button } from '@/ui/Button/Button';
import { Combobox } from '@/ui/Combobox/Combobox';
import { DateRangeField } from '@/ui/DateRangeField/DateRangeField';
import { TeamPicker } from '@/ui/TeamPicker/TeamPicker';
import type { CustomDashboardFilters } from '../model';
import { usePlanChoices, useTeamChoices } from './api';
import styles from './DashboardViewerPage.module.css';

interface FiltersBarProps {
  filters: CustomDashboardFilters;
  canReset: boolean;
  onChange: (filters: CustomDashboardFilters) => void;
  onReset: () => void;
}

export function FiltersBar({ filters, canReset, onChange, onReset }: FiltersBarProps) {
  const id = useId();
  const plans = usePlanChoices();
  const teams = useTeamChoices();
  const planChoices = plans.data ?? [];
  const known = new Set(planChoices.map((plan) => plan.id));

  return (
    <div role="group" aria-label="Dashboard filters" className={styles.filters}>
      <div className={styles.dates}>
        <DateRangeField
          label="Dates"
          idPrefix={`${id}-filter`}
          start={filters.dateFrom ?? ''}
          end={filters.dateTo ?? ''}
          onChange={({ start, end }) =>
            onChange({ ...filters, dateFrom: start || null, dateTo: end || null })
          }
        />
      </div>
      <div className={styles.picker}>
        <Combobox
          name={`${id}-plans`}
          label="Plans"
          multiple
          options={planChoices.map((plan) => ({ id: plan.id, label: plan.name }))}
          value={filters.planIds.filter((planId) => known.has(planId))}
          onChange={(planIds) => onChange({ ...filters, planIds })}
          renderChip={(planId) => ({
            label: planChoices.find((plan) => plan.id === planId)?.name ?? planId,
          })}
          placeholder="All plans"
          countText={(count) => plural(count, 'plan')}
          noMatchText={(query) => `No plan matches '${query}'.`}
        />
      </div>
      <div className={styles.picker}>
        <TeamPicker
          name={`${id}-teams`}
          label="Teams"
          teams={teams.data ?? []}
          value={filters.teamIds}
          onChange={(teamIds) => onChange({ ...filters, teamIds })}
        />
      </div>
      {canReset && (
        <Button variant="tertiary" icon={<RotateCcw />} onClick={onReset}>
          Reset filters
        </Button>
      )}
    </div>
  );
}
