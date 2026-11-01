import { Shuffle } from 'lucide-react';
import type { SettingsSection } from '@/features/settings';
import { ATTRIBUTION, ATTRIBUTION_DESCRIPTION } from './AttributionSettingsPage';

export const ATTRIBUTION_SETTINGS_SECTIONS: SettingsSection[] = [
  {
    title: ATTRIBUTION,
    description: ATTRIBUTION_DESCRIPTION,
    path: '/settings/attribution',
    icon: Shuffle,
  },
];
