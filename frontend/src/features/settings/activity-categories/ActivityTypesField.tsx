import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import type { ChoiceOption } from '@/ui/ChoiceField/ChoiceField';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import type { FieldBinding } from '@/ui/FormField/FormField';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { useActivityTypes } from '../api';
import type { ActivityCategory } from './api';

export interface ActivityTypesFieldProps extends FieldBinding<string[]> {
  /** The category being edited: the types it holds stay offered even when inactive. */
  category?: ActivityCategory;
}

/** Activity types: the active types, plus those already in this category (FR-019). */
export function ActivityTypesField({
  category,
  name,
  value,
  onChange,
  error,
}: ActivityTypesFieldProps) {
  const types = useActivityTypes();

  if (!types.data) {
    return types.isError ? (
      <ErrorState error={types.error} onRetry={() => void types.refetch()} />
    ) : (
      <Skeleton lines={3} />
    );
  }

  const options: ChoiceOption[] = types.data.items.map((type) => ({
    id: type.id,
    label: type.name,
    active: type.active,
  }));
  for (const held of category?.activityTypes ?? []) {
    if (!options.some((option) => option.id === held.id)) {
      options.push({ id: held.id, label: held.name, active: held.active });
    }
  }

  return (
    <ChoiceField
      label="Activity types"
      name={name}
      hint="A type already in another category is offered here too; choosing it moves it to this category. Types in no category count as Uncategorised."
      multiple
      options={options}
      value={value}
      onChange={onChange}
      error={error}
    />
  );
}
