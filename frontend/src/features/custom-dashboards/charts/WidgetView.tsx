import { useId, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { saveBlob } from '@/api/download';
import type { ApiError } from '@/api/errors';
import { useFormatters } from '@/format/useFormatters';
import { ActionsMenu } from '@/ui/ActionsMenu/ActionsMenu';
import type { ActionsMenuItem } from '@/ui/ActionsMenu/ActionsMenu';
import { useToast } from '@/ui/Toast/useToast';
import type { CustomDashboardFilters, Widget, WidgetData } from '../model';
import { widgetTypeInfo } from '../widget-types';
import { useWidgetData, widgetDataRequest } from './api';
import { buildOption } from './build-option';
import { describeWidget, shownGroupsNote } from './describe';
import { EChart } from './EChart';
import type { ChartHandle } from './EChart';
import { KeyFigure } from './KeyFigure';
import { pngFileName } from './png';
import { TextNote } from './TextNote';
import { useChartTheme } from './theme';
import { WidgetTable } from './WidgetTable';
import {
  WidgetError,
  WidgetGone,
  WidgetIncomplete,
  WidgetLoading,
  WidgetNoAccess,
  WidgetNoData,
} from './WidgetStates';
import styles from './WidgetView.module.css';

export interface WidgetViewProps {
  widget: Widget;
  filters: CustomDashboardFilters | null;
  mode: 'view' | 'edit';
  dashboardName: string;
  viewerOwnsDashboard: boolean;
}

const ROW_HEIGHT = 80;
const ROW_GAP = 16;
const CHROME = 72;
const NOTE_HEIGHT = 20;
const SMALLEST_CHART = 120;
const PREVIEW_DELAY_MS = 400;

export function WidgetView({
  widget,
  filters,
  mode,
  dashboardName,
  viewerOwnsDashboard,
}: WidgetViewProps) {
  const titleId = useId();
  const [isTable, setIsTable] = useState(false);
  const [handle, setHandle] = useState<ChartHandle | null>(null);
  const { show } = useToast();
  const result = useWidgetData(widget, filters, {
    debounceMs: mode === 'edit' ? PREVIEW_DELAY_MS : 0,
  });
  const data = result.data;
  const isChart = widgetTypeInfo(widget.type).group === 'charts';
  const notes = data && !result.isError ? notesOf(data) : [];
  const height = chartHeight(widget, notes.length);
  const hasRows = Boolean(data && data.rows.length > 0);

  function handleDownload() {
    if (!handle) {
      return;
    }
    handle
      .toPng()
      .then((blob) => saveBlob(blob, pngFileName(dashboardName, widget.title)))
      .catch(() =>
        show({ message: "The chart couldn't be saved as a picture. Please try again." }),
      );
  }

  const items: ActionsMenuItem[] = [
    { label: isTable ? 'Show as chart' : 'Show as table', onSelect: () => setIsTable(!isTable) },
    ...(isTable || !handle ? [] : [{ label: 'Download as PNG', onSelect: handleDownload }]),
  ];
  const showsMenu = mode === 'view' && isChart && hasRows && !result.isError;

  return (
    <section className={styles.widget} aria-labelledby={widget.title ? titleId : undefined}>
      {(widget.title || widget.subtitle || showsMenu) && (
        <header className={styles.header}>
          <div className={styles.heading}>
            {widget.title && (
              <h2 id={titleId} className={styles.title}>
                {widget.title}
              </h2>
            )}
            {widget.subtitle && <p className={styles.subtitle}>{widget.subtitle}</p>}
          </div>
          {showsMenu && <ActionsMenu label={widget.title || 'this chart'} items={items} />}
        </header>
      )}
      <div className={styles.body} aria-busy={result.isFetching && Boolean(data)}>
        <Body
          widget={widget}
          filters={filters}
          mode={mode}
          result={result}
          height={height}
          isTable={isTable}
          viewerOwnsDashboard={viewerOwnsDashboard}
          onReady={setHandle}
        />
      </div>
      {notes.length > 0 && (
        <ul className={styles.notes} role="list">
          {notes.map((note) => (
            <li key={note}>{note}</li>
          ))}
        </ul>
      )}
    </section>
  );
}

interface BodyProps extends Pick<
  WidgetViewProps,
  'widget' | 'filters' | 'mode' | 'viewerOwnsDashboard'
> {
  result: { data?: WidgetData; error: ApiError | null; refetch: () => unknown };
  height: number;
  isTable: boolean;
  onReady: (handle: ChartHandle) => void;
}

function Body({
  widget,
  filters,
  mode,
  result,
  viewerOwnsDashboard,
  ...shown
}: BodyProps): ReactNode {
  if (widget.type === 'text') {
    return <TextNote note={widget.note} isEditing={mode === 'edit'} />;
  }
  if (widgetDataRequest(widget, filters) === null) {
    return <WidgetIncomplete />;
  }
  if (result.error) {
    return failure(result.error, viewerOwnsDashboard, () => void result.refetch());
  }
  if (!result.data) {
    return <WidgetLoading height={shown.height} />;
  }
  return <Figures widget={widget} filters={filters} data={result.data} {...shown} />;
}

interface FiguresProps extends Pick<
  BodyProps,
  'widget' | 'filters' | 'height' | 'isTable' | 'onReady'
> {
  data: WidgetData;
}

function Figures({ widget, filters, data, height, isTable, onReady }: FiguresProps) {
  if (isEmpty(widget, data)) {
    const { date, plan, team } = data.filtersApplied;
    return <WidgetNoData isFiltered={date || plan || team} />;
  }
  if (widget.type === 'kpi') {
    return <KeyFigure widget={widget} data={data} filters={filters} />;
  }
  if (widget.type === 'table' || isTable) {
    return <WidgetTable widget={widget} data={data} />;
  }
  return <Chart widget={widget} data={data} height={height} onReady={onReady} />;
}

interface ChartProps {
  widget: Widget;
  data: WidgetData;
  height: number;
  onReady: (handle: ChartHandle) => void;
}

function Chart({ widget, data, height, onReady }: ChartProps) {
  const theme = useChartTheme();
  const { dateFormat } = useFormatters();
  const option = useMemo(
    () => buildOption(widget, data, theme, dateFormat),
    [widget, data, theme, dateFormat],
  );
  const description = useMemo(() => describeWidget(widget, data), [widget, data]);
  return <EChart option={option} description={description} height={height} onReady={onReady} />;
}

function failure(error: ApiError, viewerOwnsDashboard: boolean, retry: () => void) {
  if (error.kind === 'forbidden' || (error.kind === 'notFound' && !viewerOwnsDashboard)) {
    return <WidgetNoAccess />;
  }
  if (error.kind === 'notFound') {
    return <WidgetGone />;
  }
  return <WidgetError error={error} onRetry={retry} />;
}

function isEmpty(widget: Widget, data: WidgetData): boolean {
  if (widget.type === 'kpi' || widget.type === 'gauge') {
    const value = data.totals[0] ?? data.rows[0]?.values[0];
    return !value || (value.number === null && value.money === null && value.date === null);
  }
  return data.rows.length === 0;
}

function notesOf(data: WidgetData): string[] {
  const shown = shownGroupsNote(data);
  return shown ? [shown, ...data.notes] : data.notes;
}

function chartHeight(widget: Widget, noteCount: number): number {
  const cell = widget.layout.h * ROW_HEIGHT + (widget.layout.h - 1) * ROW_GAP;
  return Math.max(SMALLEST_CHART, cell - CHROME - noteCount * NOTE_HEIGHT);
}
