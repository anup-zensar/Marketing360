import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { dataSourceKeys, datasetKeys } from '../keys';
import type { Dataset, DatasetPreview } from '../model';

type S = components['schemas'];
export type DatasetList = S['DatasetList'];
export type DatasetUpload = S['DatasetUpload'];
export type DatasetCreate = S['DatasetCreate'];
export type DatasetUpdate = S['DatasetUpdate'];
export type DatasetRowPage = S['DatasetRowPage'];
export type DataFieldKind = S['DataFieldKind'];

export const ROWS_PER_PAGE = 50;

export function useDatasets() {
  return useQuery<DatasetList>({
    queryKey: datasetKeys.all,
    queryFn: ({ signal }) => call(api.GET('/api/v1/datasets', { signal })),
  });
}

export function useDataset(datasetId: string) {
  return useQuery<Dataset>({
    queryKey: datasetKeys.one(datasetId),
    queryFn: ({ signal }) =>
      call(api.GET('/api/v1/datasets/{datasetId}', { params: { path: { datasetId } }, signal })),
  });
}

export function useDatasetRows(datasetId: string, offset: number) {
  return useQuery<DatasetRowPage>({
    queryKey: datasetKeys.rows(datasetId, offset),
    queryFn: ({ signal }) =>
      call(
        api.GET('/api/v1/datasets/{datasetId}/rows', {
          params: { path: { datasetId }, query: { offset, limit: ROWS_PER_PAGE } },
          signal,
        }),
      ),
    placeholderData: keepPreviousData,
  });
}

export function usePreviewDataset() {
  return useMutation<DatasetPreview, Error, DatasetUpload>({
    mutationFn: (body) => call(api.POST('/api/v1/datasets/preview', { body })),
  });
}

export function useCreateDataset() {
  const refresh = useRefreshDatasets();
  return useMutation<Dataset, Error, DatasetCreate>({
    mutationFn: (body) => call(api.POST('/api/v1/datasets', { body })),
    onSettled: refresh,
  });
}

export function useUpdateDataset() {
  const refresh = useRefreshDatasets();
  return useMutation<Dataset, Error, { id: string; body: DatasetUpdate }>({
    mutationFn: ({ id, body }) =>
      call(
        api.PATCH('/api/v1/datasets/{datasetId}', { params: { path: { datasetId: id } }, body }),
      ),
    onSettled: refresh,
  });
}

export function useDeleteDataset() {
  const refresh = useRefreshDatasets();
  return useMutation<S['DeleteResult'], Error, string>({
    mutationFn: (datasetId) =>
      call(api.DELETE('/api/v1/datasets/{datasetId}', { params: { path: { datasetId } } })),
    onSettled: refresh,
  });
}

export function useShareDataset() {
  const refresh = useRefreshDatasets();
  return useMutation<Dataset, Error, { id: string; version: number; teamIds: string[] }>({
    mutationFn: ({ id, version, teamIds }) =>
      call(
        api.PUT('/api/v1/datasets/{datasetId}/sharing', {
          params: { path: { datasetId: id } },
          body: { version, teamIds },
        }),
      ),
    onSettled: refresh,
  });
}

function useRefreshDatasets() {
  const queryClient = useQueryClient();
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: datasetKeys.all }),
      queryClient.invalidateQueries({ queryKey: dataSourceKeys.all }),
    ]);
}
