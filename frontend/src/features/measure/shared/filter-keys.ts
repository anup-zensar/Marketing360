/** The filter panel's draft key of a Shell-attribute group. */
export const attrKey = (attributeId: string) => `attr:${attributeId}`;

/** The values ticked in a group; null when every value is (which filters nothing, FR-012). */
export type Ticked = string[] | null;

/**
 * The values of all a group restricts to: none when nothing or everything is ticked (FR-012), so a filter
 * that would keep every value is not sent.
 */
export function restrictTo<T extends string>(all: readonly T[], ticked: Ticked): T[] {
  if (ticked === null || ticked.length === 0) {
    return [];
  }
  const kept = all.filter((value) => ticked.includes(value));
  return kept.length === all.length ? [] : kept;
}
