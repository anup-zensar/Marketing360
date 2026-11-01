import { writeAddress } from '../shared/report-address';
import type { ReportAddress } from '../shared/report-address';

/** The report filter's parameters All work reads too (039 research R1). */
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

/**
 * All work with the report's filters and the one value of a widget: an attribute value ("attr=SA-12:none"), an
 * activity type or an activity status in place of that filter's own values (043 FR-014).
 */
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
