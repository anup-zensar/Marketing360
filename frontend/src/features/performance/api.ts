/** What performance data (feature 046) asks the API: Settings › Performance data. */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';

type Schemas = components['schemas'];
export type PerformanceSource = Schemas['PerformanceSource'];
export type PerformanceSourceInput = Schemas['PerformanceSourceInput'];
export type PerformanceCategory = Schemas['PerformanceCategory'];
export type PerformanceDeliveryRoute = Schemas['PerformanceDeliveryRoute'];
export type PerformanceCadence = Schemas['PerformanceCadence'];
export type PerformanceFreshness = Schemas['PerformanceFreshness'];
export type ColumnMapping = Schemas['ColumnMapping'];

export const performanceKeys = {
  sources: ['performance', 'sources'] as const,
  source: (id: string) => ['performance', 'sources', id] as const,
  mappings: (id: string) => ['performance', 'sources', id, 'mappings'] as const,
};

export function usePerformanceSources() {
  return useQuery({
    queryKey: performanceKeys.sources,
    queryFn: () => call(api.GET('/api/v1/performance/sources')),
  });
}

export function usePerformanceSource(sourceId: string) {
  return useQuery({
    queryKey: performanceKeys.source(sourceId),
    queryFn: () =>
      call(api.GET('/api/v1/performance/sources/{sourceId}', { params: { path: { sourceId } } })),
  });
}

export function useColumnMappings(sourceId: string) {
  return useQuery({
    queryKey: performanceKeys.mappings(sourceId),
    queryFn: () =>
      call(
        api.GET('/api/v1/performance/sources/{sourceId}/mappings', {
          params: { path: { sourceId } },
        }),
      ),
  });
}

export function useCreatePerformanceSource() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: PerformanceSourceInput) =>
      call(api.POST('/api/v1/performance/sources', { body })),
    onSettled: () => queryClient.invalidateQueries({ queryKey: performanceKeys.sources }),
  });
}

export function useUpdatePerformanceSource(sourceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: PerformanceSourceInput) =>
      call(
        api.PATCH('/api/v1/performance/sources/{sourceId}', {
          params: { path: { sourceId } },
          body,
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: performanceKeys.sources }),
  });
}

export function useSetSourceStatus(sourceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (next: 'deactivate' | 'reactivate') =>
      call(
        next === 'deactivate'
          ? api.POST('/api/v1/performance/sources/{sourceId}/deactivate', {
              params: { path: { sourceId } },
            })
          : api.POST('/api/v1/performance/sources/{sourceId}/reactivate', {
              params: { path: { sourceId } },
            }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: performanceKeys.sources }),
  });
}
