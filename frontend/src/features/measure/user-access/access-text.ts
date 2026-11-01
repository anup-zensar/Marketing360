import type { UserAccessRow } from './api';

/** Each role with its scope, as Settings › Roles and access › People shows them (FR-005). */
export function rolesText(row: UserAccessRow): string[] {
  return row.roles.map((a) => `${a.role.name} (${a.scope.text})`);
}

export function teamsText(row: UserAccessRow): string {
  return row.teams.map((t) => (t.active ? t.name : `${t.name} (inactive)`)).join(', ');
}

export function personHref(id: string): string {
  return `/settings/people/${encodeURIComponent(id)}`;
}
