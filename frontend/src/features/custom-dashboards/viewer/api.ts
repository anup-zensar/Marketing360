import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { dashboardKeys, widgetDataKeys } from '../keys';
import type { CustomDashboard } from '../model';

type CustomDashboardList = components['schemas']['CustomDashboardList'];
type DeleteResult = components['schemas']['DeleteResult'];

export interface Choice {
  id: string;
  name: string;
}

export function useCustomDashboards() {
  return useQuery<CustomDashboardList>({
    queryKey: dashboardKeys.all,
    queryFn: ({ signal }) => call(api.GET('/api/v1/custom-dashboards', { signal })),
  });
}

export function useCustomDashboard(dashboardId: string) {
  return useQuery<CustomDashboard>({
    queryKey: dashboardKeys.one(dashboardId),
    queryFn: ({ signal }) =>
      call(
        api.GET('/api/v1/custom-dashboards/{dashboardId}', {
          params: { path: { dashboardId } },
          signal,
        }),
      ),
  });
}

export function useDeleteCustomDashboard() {
  const queryClient = useQueryClient();
  return useMutation<DeleteResult, Error, string>({
    mutationFn: (dashboardId) =>
      call(
        api.DELETE('/api/v1/custom-dashboards/{dashboardId}', {
          params: { path: { dashboardId } },
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: dashboardKeys.all }),
  });
}

export function useCopyCustomDashboard() {
  const queryClient = useQueryClient();
  return useMutation<CustomDashboard, Error, string>({
    mutationFn: (dashboardId) =>
      call(
        api.POST('/api/v1/custom-dashboards/{dashboardId}/copy', {
          params: { path: { dashboardId } },
          body: {},
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: dashboardKeys.all }),
  });
}

export interface ShareInput {
  id: string;
  version: number;
  teamIds: string[];
}

export function useShareCustomDashboard() {
  const queryClient = useQueryClient();
  return useMutation<CustomDashboard, Error, ShareInput>({
    mutationFn: ({ id, version, teamIds }) =>
      call(
        api.PUT('/api/v1/custom-dashboards/{dashboardId}/sharing', {
          params: { path: { dashboardId: id } },
          body: { version, teamIds },
        }),
      ),
    onSuccess: (saved) => queryClient.setQueryData(dashboardKeys.one(saved.id), saved),
    onSettled: () => queryClient.invalidateQueries({ queryKey: dashboardKeys.all }),
  });
}

export function useRefreshWidgets(): () => Promise<void> {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: widgetDataKeys.all });
}

export function usePlanChoices() {
  return useQuery<Choice[]>({
    queryKey: ['plans', 'choices'],
    queryFn: async ({ signal }) => {
      const list = await call(api.GET('/api/v1/plans', { signal }));
      return list.items.map((plan) => ({ id: plan.id, name: plan.name }));
    },
  });
}

export function useTeamChoices() {
  return useQuery({
    queryKey: ['teams', 'choices'],
    queryFn: async ({ signal }) => {
      const page = await call(api.GET('/api/v1/teams', { signal }));
      return page.items.map(({ id, name, colour, active }) => ({ id, name, colour, active }));
    },
  });
}
