import { useReducer, useState } from 'react';
import { flushSync } from 'react-dom';
import { asApiError } from '@/api/errors';
import { LinkButton } from '@/ui/Button/Button';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { FailureMessage } from '@/ui/FormPanel/FormPanel';
import { useToast } from '@/ui/Toast/useToast';
import { usePageTitle } from '@/ui/usePageTitle';
import { useUnsavedChanges } from '@/ui/useUnsavedChanges';
import { useDataSources } from '../charts';
import type { CustomDashboard } from '../model';
import { SharingDialog } from '../sharing/SharingDialog';
import { useShareDashboard } from './api';
import { builderReducer, initialState, isDirty, selectedWidget } from './builder-reducer';
import { BuilderHeader } from './BuilderHeader';
import { Canvas } from './Canvas';
import { DETAIL_FIELD_IDS, NAME_FIELD_ID, widgetHandleId } from './canvas-dom';
import { Inspector } from './Inspector';
import { ProblemSummary } from './ProblemSummary';
import { useBuilderSave } from './useBuilderSave';
import { useEditorShortcuts } from './useEditorShortcuts';
import type { Problem } from './validate';
import { WidgetPalette } from './WidgetPalette';
import styles from './Builder.module.css';

export interface BuilderEditorProps {
  dashboard: CustomDashboard | null;
}

export function BuilderEditor({ dashboard }: BuilderEditorProps) {
  const [state, dispatch] = useReducer(builderReducer, dashboard, initialState);
  const [isSharing, setIsSharing] = useState(false);
  const sources = useDataSources();
  const share = useShareDashboard();
  const { show } = useToast();
  const dashboardId = dashboard?.id;
  const title = dashboard ? `Edit '${dashboard.name}'` : 'New dashboard';
  const { dialog } = useUnsavedChanges(isDirty(state), title);
  const saving = useBuilderSave({ dashboardId, state, dispatch, sources: sources.data });
  usePageTitle(title);
  useEditorShortcuts(dispatch);

  function goTo(problem: Problem) {
    flushSync(() => dispatch({ type: 'select', id: problem.widgetId ?? null }));
    const target = problem.widgetId
      ? widgetHandleId(problem.widgetId)
      : (DETAIL_FIELD_IDS[problem.field] ?? NAME_FIELD_ID);
    document.getElementById(target)?.focus();
  }

  function saveSharing(teamIds: string[]) {
    if (!dashboardId) {
      return;
    }
    share.mutate(
      { dashboardId, version: state.version, teamIds },
      {
        onSuccess: (shared) => {
          dispatch({ type: 'followVersion', version: shared.version });
          setIsSharing(false);
          show({ message: 'Sharing saved' });
        },
      },
    );
  }

  return (
    <>
      <BuilderHeader
        title={title}
        dashboardId={dashboardId}
        canShare={dashboard?.access.share.available ?? false}
        isSaving={saving.isSaving}
        onUndo={() => dispatch({ type: 'undo' })}
        onRedo={() => dispatch({ type: 'redo' })}
        onSharing={() => setIsSharing(true)}
        onSave={() => void saving.save()}
      />
      {saving.failure && <FailureMessage failure={{ message: saving.failure }} />}
      <ProblemSummary problems={saving.problems} widgets={state.draft.widgets} onGo={goTo} />
      <div className={styles.narrow}>
        <p>The builder needs a wider window.</p>
        {dashboardId && (
          <LinkButton to={`/dashboards/${dashboardId}`}>View the dashboard</LinkButton>
        )}
      </div>
      <div className={styles.editor}>
        <WidgetPalette onAdd={(widgetType) => dispatch({ type: 'add', widgetType })} />
        <Canvas
          draft={state.draft}
          selectedId={state.selectedId}
          problems={saving.problems}
          announcement={state.announcement}
          onAction={dispatch}
        />
        <Inspector
          draft={state.draft}
          selected={selectedWidget(state)}
          problems={saving.problems}
          onAction={dispatch}
        />
      </div>
      {dialog}
      <ConfirmDialog
        open={saving.hasConflict}
        onOpenChange={(open) => !open && void saving.keepEditing()}
        title="Dashboard changed elsewhere"
        cancelLabel="Keep editing"
        confirmLabel="Reload"
        onConfirm={() => void saving.reload()}
      >
        This dashboard was changed elsewhere since you opened it. Your changes are still here.
      </ConfirmDialog>
      {dashboard && (
        <SharingDialog
          open={isSharing}
          recordLabel="dashboard"
          recordName={dashboard.name}
          teams={dashboard.shareableTeams}
          selected={dashboard.sharedWith.map((team) => team.id)}
          saving={share.isPending}
          error={share.error ? asApiError(share.error).message : null}
          onSave={saveSharing}
          onClose={() => {
            setIsSharing(false);
            share.reset();
          }}
        />
      )}
    </>
  );
}
