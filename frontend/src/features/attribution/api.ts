import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';

export type AttributionSettings = components['schemas']['AttributionSettings'];
export type AttributionModel = components['schemas']['AttributionModel'];

const KEY = ['settings', 'attribution'] as const;

export function useAttributionSettings() {
  return useQuery({ queryKey: KEY, queryFn: () => call(api.GET('/api/v1/settings/attribution')) });
}

export function useSaveAttributionSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: components['schemas']['AttributionSettingsInput']) =>
      call(api.PUT('/api/v1/settings/attribution', { body })),
    onSettled: () => queryClient.invalidateQueries({ queryKey: KEY }),
  });
}
