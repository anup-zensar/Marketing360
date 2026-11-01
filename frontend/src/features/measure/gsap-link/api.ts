import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { toQuery, writeAddress } from '../shared';
import type { ReportAddress } from '../shared';
import type { GsapTable } from './gsap-address';

export type GsapLinkReport = components['schemas']['GsapLinkReport'];
export type GsapPath = components['schemas']['GsapPath'];
export type GsapBand = components['schemas']['GsapBand'];
export type GsapSwbsTab = components['schemas']['GsapSwbsTab'];
export type GsapSwbsRow = components['schemas']['GsapSwbsRow'];
export type GsapWaitingTab = components['schemas']['GsapWaitingTab'];
export type GsapWaitingRow = components['schemas']['GsapWaitingRow'];
export type GsapOrdersTab = components['schemas']['GsapOrdersTab'];
export type GsapOrderRow = components['schemas']['GsapOrderRow'];

/** One tab of the GSAP link report: its cards, chart and one page of rows (044 contracts/api.md). */
export function useGsapLink(address: ReportAddress, table: GsapTable) {
  return useQuery({
    queryKey: ['reports', 'gsapLink', writeAddress(address).toString(), table],
    queryFn: () =>
      call(
        api.GET('/api/v1/reports/gsap-link', {
          params: { query: { ...toQuery(address), ...table } },
        }),
      ),
    placeholderData: keepPreviousData,
  });
}
