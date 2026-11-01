import {
  ChartNoAxesColumn,
  CircleAlert,
  DatabaseZap,
  RotateCw,
  ShieldOff,
  SlidersHorizontal,
} from 'lucide-react';
import type { ReactNode } from 'react';
import { ApiError } from '@/api/errors';
import { Button } from '@/ui/Button/Button';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import styles from './WidgetView.module.css';

export function WidgetLoading({ height }: { height: number }) {
  return (
    <div className={styles.reserved} style={{ height }} aria-busy="true">
      <Skeleton height="100%" />
    </div>
  );
}

export function WidgetError({ error, onRetry }: { error: unknown; onRetry: () => void }) {
  const message = (error instanceof ApiError ? error : new ApiError('other')).message;
  return (
    <WidgetMessage icon={<CircleAlert />} title="This widget couldn't be loaded" text={message}>
      <Button size="sm" icon={<RotateCw />} onClick={onRetry}>
        Retry
      </Button>
    </WidgetMessage>
  );
}

export function WidgetNoData({ isFiltered }: { isFiltered: boolean }) {
  const hint = isFiltered
    ? 'Try a wider date range, other plans or teams, or fewer filters on this widget.'
    : 'Try fewer filters on this widget, or another group.';
  return <WidgetMessage icon={<ChartNoAxesColumn />} title="No data to show" text={hint} />;
}

export function WidgetNoAccess() {
  return (
    <WidgetMessage
      icon={<ShieldOff />}
      title="You don't have access to this data"
      text="Ask the dashboard's owner or an administrator if you need it."
    />
  );
}

export function WidgetGone() {
  return (
    <WidgetMessage
      icon={<DatabaseZap />}
      title="This data is no longer available"
      text="The dataset this widget used was deleted. Choose another source in the builder."
    />
  );
}

export function WidgetIncomplete() {
  return (
    <WidgetMessage
      icon={<SlidersHorizontal />}
      title="Choose what to measure"
      text="Pick a source, what to group by and a measure in the Data tab."
    />
  );
}

interface WidgetMessageProps {
  icon: ReactNode;
  title: string;
  text: string;
  children?: ReactNode;
}

function WidgetMessage({ icon, title, text, children }: WidgetMessageProps) {
  return (
    <div className={styles.message}>
      <span className={styles.messageIcon} aria-hidden="true">
        {icon}
      </span>
      <p className={styles.messageTitle}>{title}</p>
      <p className={styles.messageText}>{text}</p>
      {children}
    </div>
  );
}
