import { Star } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/ui/Button/Button';
import { SidePanel } from '@/ui/SidePanel/SidePanel';
import { useToast } from '@/ui/Toast/useToast';
import type { ReportKey } from '../reports';
import { useCreateReportView, useDeleteReportView, useUpdateReportView } from './api';
import type { ReportView } from './api';
import { RenameViewDialog } from './RenameViewDialog';
import { useViewRefusal } from './saved-views';
import styles from './SavedViews.module.css';

interface ManageViewsPanelProps {
  report: ReportKey;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  views: ReportView[];
}

/** Your saved views (FR-027, FR-028): star one as the default, rename, delete at once with Undo. */
export function ManageViewsPanel({ report, open, onOpenChange, views }: ManageViewsPanelProps) {
  const [renaming, setRenaming] = useState<ReportView | null>(null);
  const update = useUpdateReportView(report);
  const remove = useDeleteReportView(report);
  const create = useCreateReportView(report);
  const refused = useViewRefusal();
  const { show } = useToast();

  function star(view: ReportView) {
    update.mutate(
      { viewId: view.id, body: { version: view.version, isDefault: !view.isDefault } },
      {
        onSuccess: (saved) =>
          show({
            message: saved.isDefault
              ? `The report now opens on '${saved.name}'`
              : `'${saved.name}' is no longer your default view`,
          }),
        onError: (error) => refused(error, 'name'),
      },
    );
  }

  function deleteView(view: ReportView) {
    remove.mutate(view.id, {
      onSuccess: (deleted) => {
        onOpenChange(false);
        show({
          message: `View '${deleted.name}' deleted`,
          action: {
            label: 'Undo',
            altText: 'Save the view again',
            onClick: () => restore(deleted),
          },
        });
      },
      onError: (error) => refused(error, 'name'),
    });
  }

  function restore(view: ReportView) {
    create.mutate(
      { name: view.name, query: view.query, isDefault: view.isDefault },
      {
        onSuccess: (saved) => show({ message: `View '${saved.name}' restored` }),
        onError: (error) => refused(error, 'name'),
      },
    );
  }

  return (
    <>
      <SidePanel
        open={open}
        onOpenChange={onOpenChange}
        title="Your saved views"
        footer={<Button onClick={() => onOpenChange(false)}>Close</Button>}
      >
        <ul className={styles.list}>
          {views.map((view) => (
            <li key={view.id} className={styles.view}>
              <span className={styles.label}>
                {view.name}
                {view.isDefault && <span className={styles.default}>Default</span>}
              </span>
              <div className={styles.actions}>
                <Button
                  size="sm"
                  icon={<Star aria-hidden="true" />}
                  aria-pressed={view.isDefault}
                  onClick={() => star(view)}
                >
                  Default <span className="visually-hidden">{view.name}</span>
                </Button>
                <Button size="sm" onClick={() => setRenaming(view)}>
                  Rename <span className="visually-hidden">{view.name}</span>
                </Button>
                <Button size="sm" onClick={() => deleteView(view)}>
                  Delete <span className="visually-hidden">{view.name}</span>
                </Button>
              </div>
            </li>
          ))}
        </ul>
      </SidePanel>
      {renaming && (
        <RenameViewDialog
          report={report}
          view={renaming}
          views={views}
          onClose={() => setRenaming(null)}
        />
      )}
    </>
  );
}
