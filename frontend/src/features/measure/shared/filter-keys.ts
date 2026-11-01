export const attrKey = (attributeId: string) => `attr:${attributeId}`;

/** null: every value ticked, which filters nothing (FR-012). */
export type Ticked = string[] | null;

export function restrictTo<T extends string>(all: readonly T[], ticked: Ticked): T[] {
  if (ticked === null || ticked.length === 0) {
    return [];
  }
  const kept = all.filter((value) => ticked.includes(value));
  return kept.length === all.length ? [] : kept;
}
