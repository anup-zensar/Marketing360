import { WidgetView } from '../charts';
import type { CustomDashboardFilters, Widget } from '../model';
import styles from './DashboardViewerPage.module.css';

interface ViewerGridProps {
  widgets: Widget[];
  filters: CustomDashboardFilters;
  dashboardName: string;
  viewerOwnsDashboard: boolean;
}

export function ViewerGrid({
  widgets,
  filters,
  dashboardName,
  viewerOwnsDashboard,
}: ViewerGridProps) {
  return (
    <ol role="list" aria-label="Widgets" className={styles.grid}>
      {inGridOrder(widgets).map((widget) => (
        <li
          key={widget.id}
          className={styles.cell}
          style={{
            '--column': widget.layout.x + 1,
            '--row': widget.layout.y + 1,
            '--width': widget.layout.w,
            '--height': widget.layout.h,
          }}
        >
          <WidgetView
            widget={widget}
            filters={filters}
            mode="view"
            dashboardName={dashboardName}
            viewerOwnsDashboard={viewerOwnsDashboard}
          />
        </li>
      ))}
    </ol>
  );
}

// Reading order and the one-column layout of narrow windows follow rows first, then columns.
function inGridOrder(widgets: Widget[]): Widget[] {
  return widgets.toSorted((a, b) => a.layout.y - b.layout.y || a.layout.x - b.layout.x);
}
