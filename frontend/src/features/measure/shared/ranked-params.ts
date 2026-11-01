import { useSearchParams } from 'react-router';

export interface RankedState {
  sort: string;
  order: 'desc' | 'asc';
  group: string;
  size: 10 | 25 | 50;
  page: number;
  q: string;
  drill: string;
}

const SIZES = [10, 25, 50] as const;

function sizeOf(text: string | null): RankedState['size'] {
  return SIZES.find((size) => String(size) === text) ?? 10;
}

/** The ranked list's part of the address (040 FR-014); any change but the page goes back to page 1. */
export function useRankedState(defaultSort: string) {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page'));
  const state: RankedState = {
    sort: params.get('sort') ?? defaultSort,
    order: params.get('order') === 'asc' ? 'asc' : 'desc',
    group: params.get('group') ?? 'none',
    size: sizeOf(params.get('size')),
    page: Number.isInteger(page) && page > 1 ? page : 1,
    q: params.get('q') ?? '',
    drill: params.get('drill') ?? '',
  };
  const change = (patch: Partial<RankedState>) => {
    const next = new URLSearchParams(params);
    for (const [name, value] of Object.entries({ ...patch, page: patch.page ?? 1 })) {
      const unset =
        value === '' || value === 'none' || value === 1 || value === 10 || value === 'desc';
      if (unset || (name === 'sort' && value === defaultSort)) {
        next.delete(name);
      } else {
        next.set(name, String(value));
      }
    }
    setParams(next);
  };
  return { state, change };
}
