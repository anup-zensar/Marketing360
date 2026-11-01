import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import type { AccessTable } from './access-address';

export type UserAccessReport = components['schemas']['UserAccessReport'];
export type UserAccessRow = components['schemas']['UserAccessRow'];
export type UserAccessBand = components['schemas']['UserAccessBand'];
export type TeamRef = components['schemas']['TeamRef'];

/** The cards, the last sign-in bands and one page of people (045 contracts/api.md). */
export function useUserAccess(table: AccessTable) {
  return useQuery({
    queryKey: ['reports', 'userAccess', table],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/user-access', {
          params: {
            query: {
              q: table.q || undefined,
              team: table.teams,
              status: table.status,
              sort: table.sort,
              page: table.page,
              pageSize: table.pageSize,
            },
          },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}

/** Every team, for the Team filter (GET /api/v1/teams). */
export function useAccessTeams() {
  return useQuery({
    queryKey: ['reports', 'userAccess', 'teams'],
    queryFn: () => call(api.GET('/api/v1/teams')),
    select: (page): TeamRef[] =>
      page.items.map((t) => ({ id: t.id, name: t.name, colour: t.colour, active: t.active })),
  });
}
