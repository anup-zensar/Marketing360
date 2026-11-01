import { writeAddress } from '../shared';
import type { ReportAddress } from '../shared';

const ALL_WORK_PARAMS = [
  'year',
  'plan',
  'planType',
  'team',
  'owner',
  'activityType',
  'activityStatus',
  'attr',
];

export type FilterParam = 'attr' | 'activityType' | 'activityStatus';

/** All work with the report's filters (039 R1) and one value in place of its filter's ("attr=SA-12:none", FR-014). */
export function allWorkLink(
  address: ReportAddress,
  param: FilterParam,
  value: string,
  attributeId = '',
): string {
  const params = new URLSearchParams();
  for (const [key, v] of writeAddress(address)) {
    if (ALL_WORK_PARAMS.includes(key) && key !== param) {
      params.append(key, v);
    }
  }
  if (param === 'attr') {
    for (const other of writeAddress(address).getAll('attr')) {
      if (!other.startsWith(`${attributeId}:`)) {
        params.append('attr', other);
      }
    }
    params.append('attr', `${attributeId}:${value}`);
  } else {
    params.append(param, value);
  }
  return `/all-work?${params.toString()}`;
}
