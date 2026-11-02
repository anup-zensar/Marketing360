import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { toQuery, writeAddress } from '../shared';
import type { ReportAddress } from '../shared';

export type RevenueReport = components['schemas']['RevenueReport'];
export type RevenueCards = components['schemas']['RevenueCards'];
export type RevenuePeriod = components['schemas']['RevenuePeriod'];
export type RevenueFigures = components['schemas']['RevenueFigures'];
export type RevenueRanked = components['schemas']['RevenueRanked'];
export type FunnelReport = components['schemas']['FunnelReport'];
export type FunnelCards = components['schemas']['FunnelCards'];
export type FunnelPeriod = components['schemas']['FunnelPeriod'];
export type FunnelFigures = components['schemas']['FunnelFigures'];
export type FunnelRanked = components['schemas']['FunnelRanked'];

/** Both reports' own address parameters: the activity category (049-C10) and, on Funnel, the stage. */
export interface OwnParams {
  category?: string;
  stage?: 'inq' | 'mql';
}

export interface RevenueRankedParams extends OwnParams {
  sort?: 'actualRevenue' | 'projectedRevenue';
  order?: 'desc' | 'asc';
  group?: string;
  size?: 10 | 25 | 50;
  page?: number;
  q?: string;
  drill?: string;
}

export interface FunnelRankedParams extends OwnParams {
  sort?: 'actual' | 'projected';
  order?: 'desc' | 'asc';
  group?: string;
  size?: 10 | 25 | 50;
  page?: number;
  q?: string;
  drill?: string;
}

export function useRevenueReport(address: ReportAddress, own: OwnParams) {
  return useQuery({
    queryKey: ['reports', 'revenue', writeAddress(address).toString(), own],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/revenue', { params: { query: { ...toQuery(address), ...own } } }),
      ),
    placeholderData: keepPreviousData,
  });
}

export function useRevenueRanked(address: ReportAddress, ranked: RevenueRankedParams) {
  return useQuery({
    queryKey: ['reports', 'revenue', 'ranked', writeAddress(address).toString(), ranked],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/revenue/ranked', {
          params: { query: { ...toQuery(address), ...ranked } },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}

export function useFunnelReport(address: ReportAddress, own: OwnParams) {
  return useQuery({
    queryKey: ['reports', 'funnel', writeAddress(address).toString(), own],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/funnel', { params: { query: { ...toQuery(address), ...own } } }),
      ),
    placeholderData: keepPreviousData,
  });
}

export function useFunnelRanked(address: ReportAddress, ranked: FunnelRankedParams) {
  return useQuery({
    queryKey: ['reports', 'funnel', 'ranked', writeAddress(address).toString(), ranked],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/funnel/ranked', {
          params: { query: { ...toQuery(address), ...ranked } },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}
