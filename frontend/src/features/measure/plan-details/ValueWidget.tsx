import { toCents } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { ReportWidget } from '../shared';
import type { ReportAddress, ReportTableData } from '../shared';
import { allWorkLink } from './allWorkLink';
import type { PlanDetailsValue, PlanDetailsWidget } from './api';
import { BarList } from './BarList';
import type { Bar } from './BarList';
import { countsInEachText, doesNotApplyText, RECORDS, tipOf } from './plan-details-text';
import type { Level } from './plan-details-text';

const SPECIAL = ['none', 'doesNotApply', 'noDates', 'outsideYear'];
const CHART_VALUES = 10;

interface ValueWidgetProps {
  widget: PlanDetailsWidget;
  level: Level;
  address: ReportAddress;
  attributeId?: string;
  currency: string;
}

type Money = (value: PlanDetailsValue['plannedCost']) => string;

/** Records or planned cost per value: the ten largest drawn with Other, every value in the table (FR-009, FR-017). */
export function ValueWidget({ widget, level, address, attributeId, currency }: ValueWidgetProps) {
  const { money } = useFormatters();
  const href = (v: PlanDetailsValue) =>
    v.linkable && widget.filterParam
      ? allWorkLink(address, widget.filterParam, v.key, attributeId)
      : undefined;
  const note = noteOf(widget, level);
  return (
    <ReportWidget
      report="planDetails"
      widget={widget.key}
      title={widget.title}
      tip={tipOf(widget, level)}
      note={note === '' ? undefined : note}
      chart={<BarList title={widget.title} bars={barsOf(widget, money, href)} />}
      table={tableOf(widget, RECORDS[level].many, currency, money)}
    />
  );
}

/** The ten largest values, Other, then Not filled in, Does not apply and the date values. */
function barsOf(
  widget: PlanDetailsWidget,
  money: Money,
  href: (v: PlanDetailsValue) => string | undefined,
): Bar[] {
  const cost = widget.measure === 'plannedCost';
  const bar = (v: PlanDetailsValue): Bar => ({
    key: v.key,
    label: v.label,
    size: cost ? Number(toCents(v.plannedCost?.amount ?? '0')) : v.count,
    text: cost ? money(v.plannedCost) : v.count.toLocaleString('en-GB'),
    href: href(v),
    muted: SPECIAL.includes(v.key),
  });
  const regular = widget.values.filter((v) => !SPECIAL.includes(v.key));
  const drawn = widget.other ? regular.slice(0, CHART_VALUES) : regular;
  return [
    ...drawn.map(bar),
    ...(widget.other ? [bar(widget.other)] : []),
    ...widget.values.filter((v) => SPECIAL.includes(v.key)).map(bar),
  ];
}

function tableOf(
  widget: PlanDetailsWidget,
  records: string,
  currency: string,
  money: Money,
): ReportTableData {
  const cost = widget.measure === 'plannedCost';
  return {
    columns: [
      { label: 'Value' },
      { label: records, numeric: true },
      ...(cost ? [{ label: `Planned cost (${currency})`, numeric: true }] : []),
    ],
    rows: widget.values.map((v) => ({
      key: v.key,
      cells: [v.label, v.count.toLocaleString('en-GB'), ...(cost ? [money(v.plannedCost)] : [])],
    })),
    total: [
      'Total',
      widget.total.count.toLocaleString('en-GB'),
      ...(cost ? [money(widget.total.plannedCost)] : []),
    ],
  };
}

/** The per cent filled in and the records left out of it go on the filled-in widget; the others draw them as bars. */
function noteOf(widget: PlanDetailsWidget, level: Level): string {
  const filledIn = widget.measure === 'filledIn';
  const doesNotApply = widget.values.find((v) => v.key === 'doesNotApply')?.count ?? 0;
  const percent =
    widget.filledInPercent === null ? undefined : `${String(widget.filledInPercent)} % filled in`;
  return [
    filledIn && percent,
    widget.countsInEach && countsInEachText(level),
    filledIn && doesNotApplyText(doesNotApply, level),
  ]
    .filter(Boolean)
    .join(' · ');
}
