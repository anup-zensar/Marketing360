import { useSearchParams } from 'react-router';
import { FRAME_PARAMS, readAddress, writeAddress } from './report-address';
import type { ReportAddress } from './report-address';

/** Canonical address (FR-015); a report's own parameters (a tab, a sort) are kept after the frame's. */
export function useReportAddress(): [ReportAddress, (next: ReportAddress) => void] {
  const [params, setParams] = useSearchParams();
  const write = (next: ReportAddress) => {
    const written = writeAddress(next);
    for (const [name, value] of params) {
      if (!FRAME_PARAMS.includes(name)) {
        written.append(name, value);
      }
    }
    setParams(written);
  };
  return [readAddress(params), write];
}
