import type { EChartsCoreOption } from 'echarts/core';
import { DEFAULT_DATE_FORMAT } from '@/format/intl';
import type { DateFormat } from '@/format/intl';
import type { Widget, WidgetData } from '../model';
import { axisChartOption, comboOption } from './build-axis';
import { heatmapOption, scatterOption } from './build-points';
import { funnelOption, gaugeOption, pieOption } from './build-round';
import { radarOption, treemapOption } from './build-tree';
import type { ChartContext } from './chart-data';
import type { ChartTheme } from './theme';

/** The ECharts option of a chart widget, from the API's figures only; empty for the others. */
export function buildOption(
  widget: Widget,
  data: WidgetData,
  theme: ChartTheme,
  dateFormat: DateFormat = DEFAULT_DATE_FORMAT,
): EChartsCoreOption {
  const ctx: ChartContext = { widget, data, theme, display: widget.display, dateFormat };
  switch (widget.type) {
    case 'bar':
    case 'line':
    case 'area':
      return axisChartOption(ctx);
    case 'combo':
      return comboOption(ctx);
    case 'scatter':
      return scatterOption(ctx);
    case 'heatmap':
      return heatmapOption(ctx);
    case 'pie':
    case 'donut':
      return pieOption(ctx);
    case 'funnel':
      return funnelOption(ctx);
    case 'gauge':
      return gaugeOption(ctx);
    case 'treemap':
      return treemapOption(ctx);
    case 'radar':
      return radarOption(ctx);
    case 'kpi':
    case 'table':
    case 'text':
      return {};
    default: {
      const unknown: never = widget.type;
      return unknown;
    }
  }
}
