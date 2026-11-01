import { useEffect, useId, useRef, useState } from 'react';
import { Button } from '@/ui/Button/Button';
import type { ReportChoices } from './api';
import { attrKey, restrictTo } from './filter-keys';
import type { Ticked } from './filter-keys';
import { FilterGroup } from './FilterGroup';
import type { GroupOption } from './FilterGroup';
import { LevelTabs } from './LevelTabs';
import type { ReportAddress } from './report-address';
import { NOT_APPLIED, PLAN_TYPE_WORDS, STATUS_WORDS } from './report-text';
import styles from './ReportFrame.module.css';

interface Draft {
  year: number;
  ticked: Record<string, Ticked>;
}

const LIST_KEYS = [
  'plans',
  'planTypes',
  'teams',
  'owners',
  'activityTypes',
  'activityStatuses',
] as const;

type ListKey = (typeof LIST_KEYS)[number];

const TITLES: Record<ListKey, string> = {
  plans: 'Plans',
  planTypes: 'Plan type',
  teams: 'Team',
  owners: 'Owner',
  activityTypes: 'Activity type',
  activityStatuses: 'Activity status',
};

function draftOf(applied: ReportAddress, year: number): Draft {
  const ticked: Record<string, Ticked> = {};
  for (const key of LIST_KEYS) {
    ticked[key] = applied[key].length > 0 ? [...applied[key]] : null;
  }
  for (const attr of applied.attributes) {
    ticked[attrKey(attr.attributeId)] = [...attr.values];
  }
  return { year, ticked };
}

function optionsOf(choices: ReportChoices): Record<ListKey, GroupOption[]> {
  return {
    plans: choices.plans.map((p) => ({ id: p.id, label: p.name })),
    planTypes: choices.planTypes.map((t) => ({ id: t, label: PLAN_TYPE_WORDS[t] })),
    teams: [...choices.teams]
      .sort((a, b) => Number(b.mine) - Number(a.mine))
      .map((t) => ({ id: t.id, label: t.name })),
    owners: choices.owners.map((o) => ({ id: o.id, label: o.name })),
    activityTypes: choices.activityTypes.map((t) => ({ id: t.id, label: t.name })),
    activityStatuses: choices.activityStatuses.map((s) => ({ id: s, label: STATUS_WORDS[s] })),
  };
}

interface FilterPanelProps {
  choices: ReportChoices;
  applied: ReportAddress;
  year: number;
  onApply: (next: ReportAddress) => void;
}

/** The report's filters in 039 FR-008's order; they take effect on Apply (039-C1, FR-013). */
export function FilterPanel({ choices, applied, year, onApply }: FilterPanelProps) {
  const initial = draftOf(applied, year);
  const [draft, setDraft] = useState<Draft>(initial);
  const [open, setOpen] = useState(false);
  const toggle = useRef<HTMLButtonElement>(null);
  const panelId = useId();
  const pending = JSON.stringify(draft) !== JSON.stringify(initial);
  const empty = emptyGroups(draft, choices);
  const options = optionsOf(choices);
  const restricted = Object.values(initial.ticked).filter((t) => t !== null).length;
  const change = (changes: Record<string, Ticked>) =>
    setDraft((d) => ({ ...d, ticked: { ...d.ticked, ...changes } }));

  // Escape closes the collapsed panel without changing anything (FR-013).
  useEffect(() => {
    if (!open) {
      return undefined;
    }
    const close = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setDraft(draftOf(applied, year));
        setOpen(false);
        toggle.current?.focus();
      }
    };
    document.addEventListener('keydown', close);
    return () => document.removeEventListener('keydown', close);
  }, [open, applied, year]);

  function apply() {
    const ids = (key: ListKey) =>
      restrictTo(
        options[key].map((o) => o.id),
        draft.ticked[key] ?? null,
      );
    onApply({
      ...applied,
      year: draft.year,
      plans: ids('plans'),
      planTypes: restrictTo(choices.planTypes, draft.ticked.planTypes ?? null),
      teams: ids('teams'),
      owners: ids('owners'),
      activityTypes: ids('activityTypes'),
      activityStatuses: restrictTo(choices.activityStatuses, draft.ticked.activityStatuses ?? null),
      attributes: choices.attributeGroups.flatMap((g) => {
        const values = restrictTo(
          g.values.map((v) => v.id),
          draft.ticked[attrKey(g.attributeId)] ?? null,
        );
        return values.length > 0 ? [{ attributeId: g.attributeId, values }] : [];
      }),
    });
    setOpen(false);
  }

  return (
    <aside className={styles.filters} aria-label="Filters">
      <Button
        ref={toggle}
        type="button"
        variant="secondary"
        className={styles.filtersToggle}
        aria-expanded={open}
        aria-controls={panelId}
        onClick={() => setOpen(!open)}
      >
        Filters ({restricted})
      </Button>
      <div id={panelId} className={open ? styles.panelOpen : styles.panel}>
        <label className={styles.field}>
          Year
          <select
            className={styles.select}
            value={draft.year}
            aria-describedby={applied.plans.length > 0 ? `${panelId}-year` : undefined}
            onChange={(e) => onApply({ ...applied, year: Number(e.target.value), plans: [] })}
          >
            {choices.years.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </label>
        {applied.plans.length > 0 && (
          <p id={`${panelId}-year`} className={styles.pending}>
            {yearNote(applied.plans.length)}
          </p>
        )}
        {LIST_KEYS.map((key) => (
          <FilterGroup
            key={key}
            title={TITLES[key]}
            options={options[key]}
            ticked={draft.ticked[key] ?? null}
            onChange={(next) => change({ [key]: next })}
          />
        ))}
        <LevelTabs groups={choices.attributeGroups} ticked={draft.ticked} onChange={change} />
        <div className={styles.applyRow}>
          {pending && (
            <p className={styles.pending} role="status">
              {NOT_APPLIED}
            </p>
          )}
          {empty.length > 0 && (
            <p className={styles.pending} role="status">
              Tick at least one value in {empty.join(', ')}.
            </p>
          )}
          <Button type="button" variant="secondary" onClick={apply} disabled={empty.length > 0}>
            Apply
          </Button>
        </div>
      </div>
    </aside>
  );
}

/** A group with nothing ticked would match nothing, so it can't be applied (FR-013). */
function emptyGroups(draft: Draft, choices: ReportChoices): string[] {
  const lists = LIST_KEYS.filter((key) => draft.ticked[key]?.length === 0).map(
    (key) => TITLES[key],
  );
  const attributes = choices.attributeGroups
    .filter((g) => draft.ticked[attrKey(g.attributeId)]?.length === 0)
    .map((g) => g.label);
  return [...lists, ...attributes];
}

function yearNote(plans: number): string {
  return plans === 1
    ? 'Another year clears the plan you chose.'
    : `Another year clears the ${plans} plans you chose.`;
}
