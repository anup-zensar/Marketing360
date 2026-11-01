import { skipToken, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { ApiError } from '@/api/errors';
import type { components } from '@/api/generated/schema';
import { dashboardKeys } from '../keys';
import type { CustomDashboard } from '../model';

type S = components['schemas'];
export type DashboardCreate = S['CustomDashboardCreate'];
export type DashboardUpdate = S['CustomDashboardUpdate'];

export const choiceKeys = {
  plans: [...dashboardKeys.all, 'choices', 'plans'] as const,
  teams: [...dashboardKeys.all, 'choices', 'teams'] as const,
};

export interface Choice {
  id: string;
  label: string;
}

export function useDashboard(dashboardId: string | undefined) {
  return useQuery<CustomDashboard, ApiError>({
    queryKey: dashboardKeys.one(dashboardId ?? ''),
    queryFn: dashboardId ? () => fetchDashboard(dashboardId) : skipToken,
  });
}

export function useLatestDashboard() {
  const queryClient = useQueryClient();
  return (dashboardId: string) =>
    queryClient.fetchQuery({
      queryKey: dashboardKeys.one(dashboardId),
      queryFn: () => fetchDashboard(dashboardId),
      staleTime: 0,
    });
}

export function useCreateDashboard() {
  const remember = useRememberDashboard();
  return useMutation({
    mutationFn: (body: DashboardCreate) => call(api.POST('/api/v1/custom-dashboards', { body })),
    onSuccess: remember,
  });
}

export function useUpdateDashboard() {
  const remember = useRememberDashboard();
  return useMutation({
    mutationFn: ({ dashboardId, body }: { dashboardId: string; body: DashboardUpdate }) =>
      call(
        api.PATCH('/api/v1/custom-dashboards/{dashboardId}', {
          params: { path: { dashboardId } },
          body,
        }),
      ),
    onSuccess: remember,
  });
}

export function useCopyDashboard() {
  const remember = useRememberDashboard();
  return useMutation({
    mutationFn: (dashboardId: string) =>
      call(
        api.POST('/api/v1/custom-dashboards/{dashboardId}/copy', {
          params: { path: { dashboardId } },
          body: {},
        }),
      ),
    onSuccess: remember,
  });
}

export function useShareDashboard() {
  const remember = useRememberDashboard();
  return useMutation({
    mutationFn: ({ dashboardId, version, teamIds }: ShareRequest) =>
      call(
        api.PUT('/api/v1/custom-dashboards/{dashboardId}/sharing', {
          params: { path: { dashboardId } },
          body: { version, teamIds },
        }),
      ),
    onSuccess: remember,
  });
}

export function usePlanChoices() {
  return useQuery({
    queryKey: choiceKeys.plans,
    queryFn: () => call(api.GET('/api/v1/plans', { params: { query: { scope: 'all' } } })),
    select: (list): Choice[] =>
      list.items.map((plan) => ({ id: plan.id, label: `${plan.name} (${plan.year})` })),
  });
}

export function useTeamChoices() {
  return useQuery({
    queryKey: choiceKeys.teams,
    queryFn: () => call(api.GET('/api/v1/teams')),
    select: (page) => page.items,
  });
}

interface ShareRequest {
  dashboardId: string;
  version: number;
  teamIds: string[];
}

function fetchDashboard(dashboardId: string) {
  return call(
    api.GET('/api/v1/custom-dashboards/{dashboardId}', { params: { path: { dashboardId } } }),
  );
}

function useRememberDashboard() {
  const queryClient = useQueryClient();
  return (dashboard: CustomDashboard) => {
    queryClient.setQueryData(dashboardKeys.one(dashboard.id), dashboard);
    void queryClient.invalidateQueries({ queryKey: dashboardKeys.all, exact: true });
  };
}
