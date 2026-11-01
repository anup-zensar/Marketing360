import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { toQuery, writeAddress } from '../shared';
import type { ReportAddress } from '../shared';

export type PlanBudgetReport = components['schemas']['PlanBudgetReport'];
export type PlanBudgetCards = components['schemas']['PlanBudgetCards'];
export type PlanBudgetPeriod = components['schemas']['PlanBudgetPeriod'];
export type PlanBudgetFigures = components['schemas']['PlanBudgetFigures'];
export type PlanBudgetRanked = components['schemas']['PlanBudgetRanked'];

/** The ranked list's own address parameters (040 research R5). */
export interface RankedParams {
  sort?: 'budget' | 'planned' | 'committed' | 'actual' | 'leftToPlan' | 'unspent';
  order?: 'desc' | 'asc';
  group?: string;
  size?: 10 | 25 | 50;
  page?: number;
  q?: string;
  drill?: string;
}

export function usePlanBudget(address: ReportAddress) {
  return useQuery({
    queryKey: ['reports', 'planBudget', writeAddress(address).toString()],
    queryFn: () =>
      call(api.GET('/api/v1/reports/plan-budget', { params: { query: toQuery(address) } })),
    placeholderData: keepPreviousData,
  });
}

export function usePlanBudgetRanked(address: ReportAddress, ranked: RankedParams) {
  return useQuery({
    queryKey: ['reports', 'planBudget', 'ranked', writeAddress(address).toString(), ranked],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/plan-budget/ranked', {
          params: { query: { ...toQuery(address), ...ranked } },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}
