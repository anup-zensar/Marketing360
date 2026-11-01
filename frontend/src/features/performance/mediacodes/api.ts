/** Media codes (046 FR-040 - FR-042): the activity's campaign-link codes. */
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';

type Schemas = components['schemas'];
export type MediaCode = Schemas['MediaCode'];
export type MediaCodeArea = Schemas['MediaCodeArea'];
export type MediaCodeInput = Schemas['MediaCodeInput'];
export type PerformanceConfiguration = Schemas['PerformanceConfiguration'];

export const codeKeys = {
  area: (activityId: string) => ['performance', 'activity', activityId, 'media-codes'] as const,
  configuration: ['performance', 'configuration'] as const,
};

export function useMediaCodes(activityId: string) {
  return useQuery({
    queryKey: codeKeys.area(activityId),
    queryFn: () =>
      call(
        api.GET('/api/v1/activities/{activityId}/media-codes', {
          params: { path: { activityId } },
        }),
      ),
  });
}

export function usePerformanceConfiguration() {
  return useQuery({
    queryKey: codeKeys.configuration,
    queryFn: () => call(api.GET('/api/v1/performance/configuration')),
  });
}

export function useGenerateMediaCode(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: MediaCodeInput) =>
      call(
        api.POST('/api/v1/activities/{activityId}/media-codes', {
          params: { path: { activityId } },
          body,
        }),
      ),
    onSettled: () => queryClient.invalidateQueries({ queryKey: codeKeys.area(activityId) }),
  });
}

export function useDeleteMediaCode(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (codeId: string) =>
      call(api.DELETE('/api/v1/media-codes/{codeId}', { params: { path: { codeId } } })),
    onSettled: () => queryClient.invalidateQueries({ queryKey: codeKeys.area(activityId) }),
  });
}

export function useDeactivateMediaCode(activityId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (codeId: string) =>
      call(api.POST('/api/v1/media-codes/{codeId}/deactivate', { params: { path: { codeId } } })),
    onSettled: () => queryClient.invalidateQueries({ queryKey: codeKeys.area(activityId) }),
  });
}
