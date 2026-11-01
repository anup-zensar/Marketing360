import { cx } from '@/ui/cx';
import { FieldShell } from '@/ui/FieldShell/FieldShell';
import type { FieldFrameProps } from '@/ui/FieldShell/FieldShell';
import inputs from '@/ui/FieldShell/inputs.module.css';

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectGroup {
  label: string;
  options: SelectOption[];
}

export interface SelectProps extends FieldFrameProps {
  name: string;
  value: string;
  onChange: (value: string) => void;
  options?: SelectOption[];
  groups?: SelectGroup[];
  placeholder?: string;
}

export function Select({
  name,
  value,
  onChange,
  options = [],
  groups = [],
  placeholder,
  error,
  ...frame
}: SelectProps) {
  const renderOption = (option: SelectOption) => (
    <option key={option.value} value={option.value}>
      {option.label}
    </option>
  );
  return (
    <FieldShell {...frame} id={name} error={error}>
      {(describedBy, required) => (
        <select
          id={name}
          name={name}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          aria-required={required}
          className={cx(inputs.input, inputs.select, error && inputs.invalid)}
        >
          {placeholder !== undefined && <option value="">{placeholder}</option>}
          {options.map(renderOption)}
          {groups
            .filter((group) => group.options.length > 0)
            .map((group) => (
              <optgroup key={group.label} label={group.label}>
                {group.options.map(renderOption)}
              </optgroup>
            ))}
        </select>
      )}
    </FieldShell>
  );
}
