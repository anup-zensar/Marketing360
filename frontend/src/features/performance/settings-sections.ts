import { Gauge } from 'lucide-react';
import type { SettingsSection } from '@/features/settings';
import { PERFORMANCE_DATA, SETTINGS_DESCRIPTION } from './performance-text';

export const PERFORMANCE_SETTINGS_SECTIONS: SettingsSection[] = [
  {
    title: PERFORMANCE_DATA,
    description: SETTINGS_DESCRIPTION,
    path: '/settings/performance-data',
    icon: Gauge,
  },
];
