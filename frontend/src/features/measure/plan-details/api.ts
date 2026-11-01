import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { REPORTS_KEY } from '../shared/api';
import { toQuery, writeAddress } from '../shared/report-address';
import type { ReportAddress } from '../shared/report-address';

export type PlanDetailsReport = components['schemas']['PlanDetailsReport'];
export type PlanDetailsWidget = components['schemas']['PlanDetailsWidget'];
export type PlanDetailsValue = components['schemas']['PlanDetailsValue'];
export type PlanDetailsAttributeRow = components['schemas']['PlanDetailsAttributeRow'];

/** The Plan details report of the address (043 contracts/api.md). */
export function usePlanDetails(address: ReportAddress) {
  return useQuery({
    queryKey: [...REPORTS_KEY, 'planDetails', writeAddress(address).toString()],
    queryFn: () =>
      call(api.GET('/api/v1/reports/plan-details', { params: { query: toQuery(address) } })),
    placeholderData: keepPreviousData,
  });
}
