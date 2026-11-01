import type { StatusTone } from '@/ui/StatusChip/StatusChip';
import type { PerformanceCadence, PerformanceCategory, PerformanceFreshness } from './api';

export const PERFORMANCE_DATA = 'Performance data';
export const SETTINGS_DESCRIPTION =
  'Where performance figures come from, what each column of their files means, and how fresh each feed is.';

export const CATEGORY_WORDS: Record<PerformanceCategory, string> = {
  webAnalytics: 'Web analytics',
  marketingAutomation: 'Marketing automation',
  crm: 'CRM',
  leadProgression: 'Lead progression',
};

export const CADENCE_WORDS: Record<PerformanceCadence, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  monthly: 'Monthly',
  onDemand: 'On demand',
};

export const FRESHNESS_WORDS: Record<PerformanceFreshness, string> = {
  current: 'Current',
  stale: 'Stale',
  failed: 'Failed',
  inactive: 'Inactive',
};

/** Freshness is read at a glance, so its colour family matches what the word means (design system §4). */
export const FRESHNESS_TONES: Record<PerformanceFreshness, StatusTone> = {
  current: 'positive',
  stale: 'review',
  failed: 'problem',
  inactive: 'neutral',
};

export const NEVER_LOADED = 'Never';
export const EMPTY_TITLE = 'No data sources yet';
export const EMPTY_TEXT =
  'A data source says where performance figures come from. Add one to start loading data.';
