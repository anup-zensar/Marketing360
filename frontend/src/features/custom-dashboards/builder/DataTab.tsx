import { useState } from 'react';
import { Link } from 'react-router';
import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import type { DataField, DataSource, Widget, WidgetQuery } from '../model';
import { widgetTypeInfo } from '../widget-types';
import type { Need } from '../widget-types';
import { FiltersEditor } from './FiltersEditor';
import { MeasuresEditor } from './MeasuresEditor';
import { NumberInput } from './NumberInput';
import { changeSource, groupOf, withDimension } from './query-edits';
import { Select } from './Select';
import styles from './Inspector.module.css';

type GroupBy = NonNullable<WidgetQuery['dimension']>;
type DateGroup = NonNullable<GroupBy['dateGroup']>;

const DATE_GROUPS: { value: DateGroup; label: string }[] = [
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
  { value: 'quarter', label: 'Quarter' },
  { value: 'year', label: 'Year' },
];

export interface DataTabProps {
  widget: Widget;
  query: WidgetQuery;
  sources: DataSource[] | undefined;
  sourcesFailed: boolean;
  errorAt: (path: string) => string | undefined;
  onQuery: (query: WidgetQuery, typing?: string) => void;
}

export function DataTab({ widget, query, sources, sourcesFailed, errorAt, onQuery }: DataTabProps) {
  const [cleared, setCleared] = useState<string | null>(null);
  const info = widgetTypeInfo(widget.type);
  const source = sources?.find((each) => each.key === query.source);
  const prefix = `inspector-${widget.id}`;

  function chooseSource(key: string) {
    const next = sources?.find((each) => each.key === key);
    if (!next) {
      return;
    }
    const change = changeSource(query, next);
    setCleared(change.cleared);
    onQuery(change.query);
  }

  return (
    <div className={styles.tab}>
      <Select
        label="Source"
        name={`${prefix}-source`}
        marking="Required"
        value={source ? query.source : ''}
        placeholder={sources ? 'Choose a source' : 'Loading the sources…'}
        groups={[
          { label: 'SP&A One data', options: optionsOf(sources, 'builtIn') },
          { label: 'Uploaded data', options: optionsOf(sources, 'dataset') },
        ]}
        onChange={chooseSource}
        error={errorAt('query.source')}
      />
      <Link className={styles.link} to="/datasets">
        Upload data
      </Link>
      {sourcesFailed && (
        <p className={styles.error} role="alert">
          The sources could not be loaded. Reload the page to try again.
        </p>
      )}
      {cleared && (
        <p className={styles.note} role="status">
          {cleared}
        </p>
      )}
      {source && (
        <>
          <GroupPicker
            label="Group by"
            name={`${prefix}-dimension`}
            need={info.dimension}
            group={query.dimension}
            source={source}
            errorAt={(part) => errorAt(`query.dimension${part}`)}
            onChange={(dimension, field) =>
              onQuery(withDimension(widget.type, query, dimension, field))
            }
          />
          <GroupPicker
            label="Break down by"
            name={`${prefix}-series`}
            need={info.series}
            group={query.series}
            source={source}
            errorAt={(part) => errorAt(`query.series${part}`)}
            onChange={(series) => onQuery({ ...query, series })}
          />
          {info.measures[1] > 0 && (
            <MeasuresEditor
              idPrefix={prefix}
              measures={query.measures}
              max={info.measures[1]}
              source={source}
              errorAt={errorAt}
              onChange={(measures, typing) => onQuery({ ...query, measures }, typing)}
            />
          )}
          <FiltersEditor
            idPrefix={prefix}
            filters={query.filters}
            source={source}
            errorAt={errorAt}
            onChange={(filters, typing) => onQuery({ ...query, filters }, typing)}
          />
          {query.dimension && (
            <SortAndLimit prefix={prefix} query={query} errorAt={errorAt} onQuery={onQuery} />
          )}
        </>
      )}
    </div>
  );
}

function optionsOf(sources: DataSource[] | undefined, kind: DataSource['kind']) {
  return (sources ?? [])
    .filter((source) => source.kind === kind)
    .map((source) => ({ value: source.key, label: source.label }));
}

interface GroupPickerProps {
  label: string;
  name: string;
  need: Need;
  group: GroupBy | null;
  source: DataSource;
  errorAt: (part: string) => string | undefined;
  onChange: (group: GroupBy | null, field: DataField | undefined) => void;
}

function GroupPicker({ label, name, need, group, source, errorAt, onChange }: GroupPickerProps) {
  if (need === 'none') {
    return null;
  }
  const groupable = source.fields.filter((field) => field.groupable);
  const field = groupable.find((each) => each.key === group?.field);
  return (
    <>
      <ChoiceField
        label={label}
        name={name}
        marking={need === 'required' ? 'Required' : undefined}
        options={groupable.map((each) => ({ id: each.key, label: each.label, active: true }))}
        value={field ? field.key : ''}
        onChange={(key) => {
          const chosen = groupable.find((each) => each.key === key);
          onChange(groupOf(chosen), chosen);
        }}
        error={errorAt('')}
      />
      {group && field?.kind === 'date' && (
        <Select
          label={`Group dates by (${label.toLowerCase()})`}
          name={`${name}-dates`}
          value={group.dateGroup ?? ''}
          placeholder="Choose…"
          options={DATE_GROUPS}
          onChange={(value) => {
            const dateGroup = DATE_GROUPS.find((each) => each.value === value)?.value ?? null;
            onChange({ ...group, dateGroup }, field);
          }}
          error={errorAt('.dateGroup')}
        />
      )}
    </>
  );
}

interface SortAndLimitProps {
  prefix: string;
  query: WidgetQuery;
  errorAt: (path: string) => string | undefined;
  onQuery: (query: WidgetQuery, typing?: string) => void;
}

function SortAndLimit({ prefix, query, errorAt, onQuery }: SortAndLimitProps) {
  const { sort } = query;
  const sortValue = sort.by === 'dimension' ? 'dimension' : `m${sort.measureIndex}`;
  const measureOptions = query.measures.map((measure, index) => ({
    value: `m${index}`,
    label: measure.label || `Measure ${index + 1}`,
  }));
  return (
    <>
      <div className={styles.pair}>
        <Select
          label="Sort by"
          name={`${prefix}-sort`}
          value={sortValue}
          options={[{ value: 'dimension', label: 'The group' }, ...measureOptions]}
          onChange={(value) =>
            onQuery({
              ...query,
              sort:
                value === 'dimension'
                  ? { ...sort, by: 'dimension', measureIndex: 0 }
                  : { ...sort, by: 'measure', measureIndex: Number(value.slice(1)) },
            })
          }
          error={errorAt('query.sort')}
        />
        <Select
          label="Order"
          name={`${prefix}-direction`}
          value={sort.direction}
          options={[
            { value: 'asc', label: 'Ascending' },
            { value: 'desc', label: 'Descending' },
          ]}
          onChange={(value) =>
            onQuery({ ...query, sort: { ...sort, direction: value === 'asc' ? 'asc' : 'desc' } })
          }
        />
      </div>
      <NumberInput
        label="Show top"
        name={`${prefix}-limit`}
        value={query.limit}
        min={1}
        max={500}
        error={errorAt('query.limit')}
        onChange={(limit) => onQuery({ ...query, limit }, 'limit')}
      />
      <CheckboxField
        label="Group the rest as Other"
        name={`${prefix}-group-rest`}
        checked={query.groupRest}
        onChange={(groupRest) => onQuery({ ...query, groupRest })}
      />
    </>
  );
}
