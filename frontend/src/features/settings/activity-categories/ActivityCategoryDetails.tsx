import { Pencil } from 'lucide-react';
import { ActionsMenu } from '@/ui/ActionsMenu/ActionsMenu';
import { Button } from '@/ui/Button/Button';
import { SidePanel } from '@/ui/SidePanel/SidePanel';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import styles from '@/ui/settings-table.module.css';
import { CanChange } from '@/ui/CanChange/CanChange';
import { MetaLine } from '../MetaLine';
import { SETTING_NOUNS } from '../settings-nouns';
import { UsageLine } from '../UsageLine';
import { moreActions } from './category-actions';
import type { ActivityCategoryActions } from './category-actions';
import type { ActivityCategory } from './api';

export interface ActivityCategoryDetailsProps {
  /** The activity category; null when it no longer exists. */
  category: ActivityCategory | null;
  actions: ActivityCategoryActions;
  onClose: () => void;
}

/**
 * An activity category's details in a side panel (US3): whether the reports open on it, the
 * activity types it groups, where it is used and who created and last changed it; **Edit** and ⋯
 * (Set inactive or Reactivate, Delete).
 */
export function ActivityCategoryDetails({
  category,
  actions,
  onClose,
}: ActivityCategoryDetailsProps) {
  if (!category) {
    return (
      <SidePanel
        open
        onOpenChange={(open) => !open && onClose()}
        title="Activity category"
        footer={<Button onClick={onClose}>Back to Activity categories</Button>}
      >
        <p role="alert" className={styles.gone}>
          {SETTING_NOUNS.activityCategory.notFound}
        </p>
      </SidePanel>
    );
  }

  return (
    <SidePanel
      open
      onOpenChange={(open) => !open && onClose()}
      title={category.name}
      footer={
        <CanChange otherwise={<Button onClick={onClose}>Close</Button>}>
          <>
            <Button variant="primary" icon={<Pencil />} onClick={() => actions.edit(category)}>
              Edit
            </Button>
            <ActionsMenu label={category.name} items={moreActions(category, actions)} />
          </>
        </CanChange>
      }
    >
      <div className={styles.details}>
        <div className={styles.statusRow}>
          <StatusChip tone={category.active ? 'positive' : 'neutral'}>
            {category.active ? 'Active' : 'Inactive'}
          </StatusChip>
          {category.isDefault && <StatusChip tone="active">Default</StatusChip>}
          <span>{category.id}</span>
        </div>
        <dl className={styles.facts}>
          <div>
            <dt>Default category</dt>
            <dd>{category.isDefault ? 'The reports open on this category.' : 'No'}</dd>
          </div>
          <div>
            <dt>Activity types</dt>
            <dd>
              {category.activityTypes.length === 0 ? (
                '—'
              ) : (
                <ul role="list" className={styles.chips}>
                  {category.activityTypes.map((type) => (
                    <li key={type.id} className={styles.chip}>
                      {type.active ? type.name : `${type.name} (inactive)`}
                    </li>
                  ))}
                </ul>
              )}
            </dd>
          </div>
          <div>
            <dt>Used by</dt>
            <dd>
              <UsageLine usage={category.usage} />
            </dd>
          </div>
        </dl>
        <MetaLine record={category} />
      </div>
    </SidePanel>
  );
}
