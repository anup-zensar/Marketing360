import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import type { FieldBinding } from '@/ui/FormField/FormField';
import type { GoalStage } from './schemas';

const OPTIONS = [
  { id: 'inq', label: 'INQ', active: true },
  { id: 'mql', label: 'MQL', active: true },
  { id: 'revenue', label: 'Revenue', active: true },
];

/** The stages a goal may be set against; a projection runs forward from there (047 FR-002). */
export function GoalStagesField({ name, value, onChange, error }: FieldBinding<GoalStage[]>) {
  return (
    <ChoiceField
      label="Allowed goal stages"
      name={name}
      marking="Required"
      hint="A goal can be set against these stages; a projection runs forward from there."
      multiple
      options={OPTIONS}
      value={value}
      onChange={(chosen) => onChange(chosen.filter(isGoalStage))}
      error={error}
    />
  );
}

function isGoalStage(id: string): id is GoalStage {
  return OPTIONS.some((option) => option.id === id);
}
