import type { EChartsCoreOption, EChartsType } from 'echarts/core';
import { useEffect, useRef } from 'react';
import styles from './EChart.module.css';
import { createChart } from './echarts';
import { svgToPng } from './png';
import { useChartTheme } from './theme';

export interface ChartHandle {
  toPng(): Promise<Blob>;
}

export interface EChartProps {
  option: EChartsCoreOption;
  description: string;
  height: number;
  onReady?: (handle: ChartHandle) => void;
}

const REDUCED_MOTION = '(prefers-reduced-motion: reduce)';

export function EChart({ option, description, height, onReady }: EChartProps) {
  const element = useRef<HTMLDivElement>(null);
  const chart = useRef<EChartsType | null>(null);
  const { surface } = useChartTheme();

  useEffect(() => {
    const container = element.current;
    if (!container) {
      return undefined;
    }
    const created = createChart(container);
    chart.current = created;
    const observer = new ResizeObserver(() => created.resize());
    observer.observe(container);
    return () => {
      observer.disconnect();
      created.dispose();
      chart.current = null;
    };
  }, []);

  useEffect(() => {
    chart.current?.setOption(
      {
        ...option,
        animation: !window.matchMedia(REDUCED_MOTION).matches,
        aria: { enabled: true, label: { enabled: true, description }, decal: { show: true } },
      },
      { notMerge: true },
    );
  }, [option, description]);

  useEffect(() => {
    const drawn = chart.current;
    if (!onReady || !drawn) {
      return;
    }
    onReady({
      toPng: () =>
        svgToPng({
          svgUrl: drawn.getDataURL({ type: 'svg' }),
          width: drawn.getWidth(),
          height: drawn.getHeight(),
          background: surface,
        }),
    });
  }, [onReady, surface]);

  return (
    <div
      ref={element}
      role="img"
      aria-label={description}
      className={styles.chart}
      style={{ height }}
    />
  );
}
