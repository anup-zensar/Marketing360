import { z } from 'zod';

export const NAME_MESSAGE = 'Give the data source a name of 1 to 120 characters.';
export const OWNER_MESSAGE = 'Choose an owner for this data source.';

/** A data source as the form holds it; the API applies the same rules (046 FR-001). */
export const sourceSchema = z.object({
  name: z.string().trim().min(1, NAME_MESSAGE).max(120, NAME_MESSAGE),
  description: z.string(),
  ownerId: z.string().min(1, OWNER_MESSAGE),
  category: z.enum(['webAnalytics', 'marketingAutomation', 'crm', 'leadProgression']),
  deliveryRoute: z.enum(['upload', 'simulated']),
  cadence: z.enum(['daily', 'weekly', 'monthly', 'onDemand']),
  integrationId: z.string(),
});

export type SourceValues = z.infer<typeof sourceSchema>;
