import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { toQuery, writeAddress } from '../shared';
import type { ReportAddress } from '../shared';
import type { ApprovalsParams } from './schemas';

export type ApprovalsReportData = components['schemas']['ApprovalsReport'];
export type StatusMix = ApprovalsReportData['statusMix'];
export type Average = ApprovalsReportData['average'];

export function useApprovalsReport(address: ReportAddress, params: ApprovalsParams) {
  return useQuery({
    queryKey: ['reports', 'approvals', writeAddress(address).toString(), params],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/approvals', {
          params: { query: { ...toQuery(address), statusType: params.statusType } },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}
