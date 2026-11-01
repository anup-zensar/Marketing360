import { LayoutDashboard } from 'lucide-react';
import { Link, useNavigate } from 'react-router';
import { plural } from '@/format/text';
import { useFormatters } from '@/format/useFormatters';
import { isShown } from '@/session/access';
import type { Availability } from '@/session/access';
import { ActionsMenu } from '@/ui/ActionsMenu/ActionsMenu';
import type { ActionsMenuItem } from '@/ui/ActionsMenu/ActionsMenu';
import { TeamPills } from '@/ui/TeamPill/TeamPill';
import type { CustomDashboardSummary } from '../model';
import styles from './DashboardListPage.module.css';

interface DashboardCardProps {
  dashboard: CustomDashboardSummary;
  canBuild: boolean;
  onCopy: (id: string) => void;
  onDelete: (dashboard: CustomDashboardSummary) => void;
}

export function DashboardCard({ dashboard, canBuild, onCopy, onDelete }: DashboardCardProps) {
  const navigate = useNavigate();
  const { dateTime } = useFormatters();
  const href = `/dashboards/${encodeURIComponent(dashboard.id)}`;
  const items = menuItems(dashboard, canBuild, {
    open: () => void navigate(href),
    edit: () => void navigate(`/builder/${encodeURIComponent(dashboard.id)}`),
    copy: () => onCopy(dashboard.id),
    remove: () => onDelete(dashboard),
  });

  return (
    <li className={styles.card}>
      <div className={styles.cardHeader}>
        <h3 className={styles.name}>
          <Link to={href}>{dashboard.name}</Link>
        </h3>
        <ActionsMenu label={dashboard.name} items={items} />
      </div>
      {dashboard.description && <p className={styles.description}>{dashboard.description}</p>}
      <p className={styles.facts}>
        {`By ${dashboard.owner.name} · Updated ${dateTime(dashboard.updatedAt)} · ${plural(dashboard.widgetCount, 'widget')}`}
      </p>
      {dashboard.sharedWith.length > 0 ? (
        <TeamPills teams={dashboard.sharedWith} />
      ) : (
        <p className={styles.facts}>Not shared</p>
      )}
    </li>
  );
}

interface MenuActions {
  open: () => void;
  edit: () => void;
  copy: () => void;
  remove: () => void;
}

function menuItems(
  { access }: CustomDashboardSummary,
  canBuild: boolean,
  actions: MenuActions,
): ActionsMenuItem[] {
  const items: ActionsMenuItem[] = [{ label: 'Open', onSelect: actions.open }];
  if (isShown(access.edit)) {
    items.push({ label: 'Edit', onSelect: actions.edit, ...availability(access.edit) });
  }
  if (canBuild && isShown(access.copy)) {
    items.push({ label: 'Make a copy', onSelect: actions.copy, ...availability(access.copy) });
  }
  if (isShown(access.delete)) {
    items.push({
      label: 'Delete',
      tone: 'danger',
      separatorBefore: true,
      onSelect: actions.remove,
      ...availability(access.delete),
    });
  }
  return items;
}

function availability(action: Availability) {
  return { disabled: !action.available, hint: action.reason || undefined };
}

const BUILT_IN = [
  {
    kind: 'executive',
    name: 'Executive dashboard',
    text: 'Budget, planned cost and spend of every market you may see',
  },
  {
    kind: 'regional',
    name: 'Regional dashboard',
    text: 'The same for one market, with its plans and campaigns',
  },
] as const;

export function BuiltInCards() {
  return (
    <ul role="list" className={styles.cards}>
      {BUILT_IN.map((card) => (
        <li key={card.kind} className={styles.card}>
          <div className={styles.cardHeader}>
            <h3 className={styles.name}>
              <Link to={`/dashboards/${card.kind}`}>{card.name}</Link>
            </h3>
            <LayoutDashboard aria-hidden="true" className={styles.icon} />
          </div>
          <p className={styles.description}>{card.text}</p>
        </li>
      ))}
    </ul>
  );
}
