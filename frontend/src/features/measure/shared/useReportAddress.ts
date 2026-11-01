import { useSearchParams } from 'react-router';
import { readAddress, writeAddress } from './report-address';
import type { ReportAddress } from './report-address';

/** The report's address (FR-015): read with malformed values dropped, written in canonical form. */
export function useReportAddress(): [ReportAddress, (next: ReportAddress) => void] {
  const [params, setParams] = useSearchParams();
  return [readAddress(params), (next) => setParams(writeAddress(next))];
}
