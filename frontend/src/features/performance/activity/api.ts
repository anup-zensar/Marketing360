/** The activity's Data sources area (046 FR-031): keys in, figures out. */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';

type Schemas = components['schemas'];
export type DataSourcesArea = Schemas['DataSourcesArea'];
export type ExternalNameAlias = Schemas['ExternalNameAlias'];
export type SourceOfFacts = Schemas['SourceOfFacts'];
export type IntegrationId = Schemas['IntegrationId'];

export const areaKeys = {
  area: (activityId: string) => ['performance', 'activity', activityId, 'sources'] as const,
};

export function useDataSourcesArea(activityId: string) {
  return useQuery({
    queryKey: areaKeys.area(activityId),
    queryFn: () =>
      call(
        api.GET('/api/v1/activities/{activityId}/performance/sources', {
          params: { path: { activityId } },
        }),
      ),
  });
}

export function useSetIntegrationId(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (key: { sourceId: string; value: string }) =>
      call(
        api.PUT('/api/v1/activities/{activityId}/integration-ids/{sourceId}', {
          params: { path: { activityId, sourceId: key.sourceId } },
          body: { value: key.value },
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: areaKeys.area(activityId) }),
  });
}

export function useDeleteIntegrationId(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (sourceId: string) =>
      call(
        api.DELETE('/api/v1/activities/{activityId}/integration-ids/{sourceId}', {
          params: { path: { activityId, sourceId } },
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: areaKeys.area(activityId) }),
  });
}

export function useDeleteAlias(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (aliasId: string) =>
      call(
        api.DELETE('/api/v1/activities/{activityId}/external-name-aliases/{aliasId}', {
          params: { path: { activityId, aliasId } },
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: areaKeys.area(activityId) }),
  });
}

export type PerformancePanel = Schemas2['PerformancePanel'];
export type StageFigure = Schemas2['StageFigure'];
export type RevenueFigure = Schemas2['RevenueFigure'];
export type MetricFigure = Schemas2['MetricFigure'];
export type PanelSource = Schemas2['PanelSource'];

type Schemas2 = components['schemas'];

export const panelKeys = {
  panel: (activityId: string) => ['performance', 'activity', activityId, 'panel'] as const,
};

export function usePerformancePanel(activityId: string) {
  return useQuery({
    queryKey: panelKeys.panel(activityId),
    queryFn: () =>
      call(
        api.GET('/api/v1/activities/{activityId}/performance', {
          params: { path: { activityId } },
        }),
      ),
  });
}
