import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { toQuery, writeAddress } from '../shared';
import type { ReportAddress } from '../shared';

export type PlanDetailsReport = components['schemas']['PlanDetailsReport'];
export type PlanDetailsWidget = components['schemas']['PlanDetailsWidget'];
export type PlanDetailsValue = components['schemas']['PlanDetailsValue'];
export type PlanDetailsAttributeRow = components['schemas']['PlanDetailsAttributeRow'];

/** The Plan details report of the address (043 contracts/api.md). */
export function usePlanDetails(address: ReportAddress) {
  return useQuery({
    queryKey: ['reports', 'planDetails', writeAddress(address).toString()],
    queryFn: () =>
      call(api.GET('/api/v1/reports/plan-details', { params: { query: toQuery(address) } })),
    placeholderData: keepPreviousData,
  });
}
