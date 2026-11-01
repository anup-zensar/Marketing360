import { z } from 'zod';

/** The model form's values, as text exactly as typed; the API checks them (FR-009, FR-010). */
export const modelSchema = z.object({
  name: z.string(),
  inqToMqlRate: z.string(),
  mqlToDealRate: z.string(),
  averageDealSize: z.string(),
  activityTypeIds: z.array(z.string()),
  inqToMqlVelocityMonths: z.string(),
  mqlToDealVelocityMonths: z.string(),
  allowedGoalStages: z.array(z.enum(['inq', 'mql', 'revenue'])),
});

export type ModelValues = z.infer<typeof modelSchema>;

export type GoalStage = ModelValues['allowedGoalStages'][number];
