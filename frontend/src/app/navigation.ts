import {
  ChartNoAxesCombined,
  CircleCheckBig,
  ClipboardList,
  FileText,
  House,
  Landmark,
  LayoutDashboard,
  LayoutGrid,
  ReceiptText,
  RotateCcwClock,
  Settings,
  Trash,
  UserRound,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { lazy } from 'react';
import type { ComponentType, ReactNode } from 'react';
import type { Permission } from '@/session/access';

// The badge loads with the approvals code after the frame is drawn, off the first page.
const ApprovalsNavBadge = lazy(() =>
  import('@/features/approvals').then((feature) => ({ default: feature.ApprovalsNavBadge })),
);

const PurchaseSupportOnly = lazy(() =>
  import('@/features/purchase-requests').then((feature) => ({
    default: feature.PurchaseSupportOnly,
  })),
);

export interface NavItem {
  label: string;
  path: string;
  icon: LucideIcon;
  Badge?: ComponentType<{ collapsed: boolean }>;
  /** Other address beginnings under which the section is current too (Plans: line items). */
  alsoCurrentOn?: string[];
  /** Shown only to people who hold one of them; the address checks it too (RequirePermission). */
  permissions?: Permission[];
  /** Wraps the item, which it may leave out (Purchase requests while Purchase Support is off). */
  Shown?: ComponentType<{ children: ReactNode }>;
}

export interface ComingSoonSection extends NavItem {
  comingSoon: string;
}

const HOME: NavItem = { label: 'Home', path: '/', icon: House };

const PLANS: NavItem = {
  label: 'Plans',
  path: '/plans',
  icon: ClipboardList,
  alsoCurrentOn: ['/activities/', '/line-items/'],
};

const GLOBAL_BUDGETS: NavItem = {
  label: 'Global budgets',
  path: '/budgets',
  icon: Landmark,
  permissions: ['globalBudgets.view'],
};

const APPROVALS: NavItem = {
  label: 'Approvals',
  path: '/approvals',
  icon: CircleCheckBig,
  Badge: ApprovalsNavBadge,
};

const ALL_WORK: NavItem = { label: 'All work', path: '/all-work', icon: LayoutGrid };

const DASHBOARDS: NavItem = {
  label: 'Dashboards',
  path: '/dashboards',
  icon: LayoutDashboard,
  permissions: ['plans.view', 'dashboards.view'],
};

const DASHBOARD_BUILDER: NavItem = {
  label: 'Dashboard builder',
  path: '/builder',
  icon: ChartNoAxesCombined,
  alsoCurrentOn: ['/datasets'],
  permissions: ['dashboards.build'],
};

const FINANCIALS: NavItem = {
  label: 'Financials',
  path: '/financials',
  icon: ReceiptText,
  permissions: ['finance.view'],
};

const PURCHASE_REQUESTS: NavItem = {
  label: 'Purchase requests',
  path: '/purchase-requests',
  icon: FileText,
  permissions: ['purchaseRequests.list'],
  Shown: PurchaseSupportOnly,
};

const SETTINGS: NavItem = { label: 'Settings', path: '/settings', icon: Settings };

const RECYCLE_BIN: NavItem = { label: 'Recycle bin', path: '/recycle-bin', icon: Trash };

const ACTIVITY_LOG: NavItem = {
  label: 'Activity log',
  path: '/activity-log',
  icon: RotateCcwClock,
};

/** My profile: reached from the account menu, not from the sidebar. */
export const MY_PROFILE: NavItem = { label: 'My profile', path: '/profile', icon: UserRound };

export const MAIN_ITEMS: NavItem[] = [
  HOME,
  PLANS,
  GLOBAL_BUDGETS,
  APPROVALS,
  ALL_WORK,
  FINANCIALS,
  PURCHASE_REQUESTS,
  DASHBOARDS,
  DASHBOARD_BUILDER,
];

export const MORE_ITEMS: NavItem[] = [SETTINGS, RECYCLE_BIN, ACTIVITY_LOG];

export const COMING_SOON_SECTIONS: ComingSoonSection[] = [];

export function navItemsFor(items: NavItem[], has: (permission: Permission) => boolean): NavItem[] {
  return items.filter((item) => !item.permissions || item.permissions.some(has));
}
