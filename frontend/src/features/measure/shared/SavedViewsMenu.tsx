import { Bookmark, Check } from 'lucide-react';
import { DropdownMenu } from 'radix-ui';
import { useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router';
import { Button } from '@/ui/Button/Button';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { useToast } from '@/ui/Toast/useToast';
import type { ReportKey } from '../reports';
import { useReportViews, useUpdateReportView } from './api';
import type { ReportView } from './api';
import { ManageViewsPanel } from './ManageViewsPanel';
import { SaveViewDialog } from './SaveViewDialog';
import { sameQuery, useViewRefusal } from './saved-views';
import styles from './SavedViews.module.css';

type Dialog = 'save' | 'update' | 'manage' | null;

/**
 * Saved views (FR-024 – FR-029): the person's views of the report, the active one ticked, "Modified" when the
 * address differs from it. A view is an address, so applying one is a navigation and Back returns.
 */
export function SavedViewsMenu({ report }: { report: ReportKey }) {
  const [params, setParams] = useSearchParams();
  const query = params.toString();
  const views = useReportViews(report).data?.items;
  const [appliedId, setAppliedId] = useState<string | null>(null);
  const [dialog, setDialog] = useState<Dialog>(null);
  const chosen = useRef<(() => void) | null>(null);
  const update = useUpdateReportView(report);
  const refused = useViewRefusal();
  const { show } = useToast();
  const matched = views?.find((v) => sameQuery(v.query, query));
  if (matched && matched.id !== appliedId) {
    setAppliedId(matched.id);
  }
  const active = views?.find((v) => v.id === (matched?.id ?? appliedId)) ?? null;
  const modified = active !== null && !sameQuery(active.query, query);
  useDefaultView(views, query, (view) =>
    setParams(new URLSearchParams(view.query), { replace: true }),
  );

  function apply(view: ReportView) {
    setAppliedId(view.id);
    setParams(new URLSearchParams(view.query));
  }

  function updateActive(view: ReportView) {
    update.mutate(
      { viewId: view.id, body: { version: view.version, query } },
      {
        onSuccess: (saved) => show({ message: `View '${saved.name}' updated` }),
        onError: (error) => refused(error, 'name'),
        onSettled: () => setDialog(null),
      },
    );
  }

  const item = (label: string, action: () => void) => (
    <DropdownMenu.Item
      className={styles.item}
      onSelect={() => {
        chosen.current = action;
      }}
    >
      {label}
    </DropdownMenu.Item>
  );

  return (
    <>
      <DropdownMenu.Root modal={false}>
        <DropdownMenu.Trigger asChild>
          <Button icon={<Bookmark />}>
            {active ? `View: ${active.name}` : 'Saved views'}
            {modified && ' '}
            {modified && <span className={styles.modified}>Modified</span>}
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content
            className={styles.menu}
            align="end"
            sideOffset={4}
            onCloseAutoFocus={() => {
              const action = chosen.current;
              chosen.current = null;
              action?.();
            }}
          >
            {views?.length === 0 && (
              <p className={styles.empty}>No saved views yet — set the filters, then save them.</p>
            )}
            {views?.map((view) => (
              <DropdownMenu.CheckboxItem
                key={view.id}
                className={styles.item}
                checked={view.id === active?.id}
                onSelect={() => {
                  chosen.current = () => apply(view);
                }}
              >
                <span className={styles.tick}>
                  <DropdownMenu.ItemIndicator>
                    <Check aria-hidden="true" />
                  </DropdownMenu.ItemIndicator>
                </span>
                <span className={styles.name}>{view.name}</span>
                {view.isDefault && <span className={styles.default}>Default</span>}
              </DropdownMenu.CheckboxItem>
            ))}
            {views && views.length > 0 && <DropdownMenu.Separator className={styles.separator} />}
            {item(modified ? 'Save as new view…' : 'Save current filters…', () =>
              setDialog('save'),
            )}
            {modified && item('Update view…', () => setDialog('update'))}
            {views && views.length > 0 && item('Manage views…', () => setDialog('manage'))}
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
      {dialog === 'save' && (
        <SaveViewDialog
          report={report}
          views={views ?? []}
          query={query}
          onOpenChange={(open) => !open && setDialog(null)}
          onSaved={(view) => setAppliedId(view.id)}
        />
      )}
      {dialog === 'update' && active && (
        <ConfirmDialog
          open
          onOpenChange={(open) => !open && setDialog(null)}
          title={`Update your view '${active.name}'?`}
          confirmLabel="Update view"
          onConfirm={() => updateActive(active)}
          busy={update.isPending}
        >
          It will keep the filters you see now instead of the ones it had.
        </ConfirmDialog>
      )}
      <ManageViewsPanel
        report={report}
        open={dialog === 'manage'}
        onOpenChange={(open) => !open && setDialog(null)}
        views={views ?? []}
      />
    </>
  );
}

/** Opening the report with nothing in its address applies the default view, once per visit (FR-027). */
function useDefaultView(
  views: ReportView[] | undefined,
  query: string,
  apply: (view: ReportView) => void,
) {
  const [opened] = useState(query);
  const checked = useRef(false);
  useEffect(() => {
    if (checked.current || !views) {
      return;
    }
    checked.current = true;
    const preferred = views.find((v) => v.isDefault);
    if (opened === '' && query === '' && preferred) {
      apply(preferred);
    }
  }, [views, opened, query, apply]);
}
