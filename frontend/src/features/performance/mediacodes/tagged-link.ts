import type { MediaCode } from './api';

/** The code appended to the landing page, the way the outside tool will send it back (FR-042). */
export function taggedLink(code: MediaCode): string {
  if (!code.landingPage) {
    return String(code.code);
  }
  const glue = code.landingPage.includes('?') ? '&' : '?';
  return `${code.landingPage}${glue}mc=${code.code}`;
}
