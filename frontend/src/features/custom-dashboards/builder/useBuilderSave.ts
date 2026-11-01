import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { asApiError } from '@/api/errors';
import type { ApiError } from '@/api/errors';
import { useToast } from '@/ui/Toast/useToast';
import type { DataSource, Widget } from '../model';
import { useCreateDashboard, useLatestDashboard, useUpdateDashboard } from './api';
import type { BuilderAction, BuilderState } from './builder-reducer';
import type { Problem } from './validate';
import { validateDraft } from './validate';

interface SaveOptions {
  dashboardId: string | undefined;
  state: BuilderState;
  dispatch: (action: BuilderAction) => void;
  sources: DataSource[] | undefined;
}

export function useBuilderSave({ dashboardId, state, dispatch, sources }: SaveOptions) {
  const navigate = useNavigate();
  const { show } = useToast();
  const create = useCreateDashboard();
  const update = useUpdateDashboard();
  const latest = useLatestDashboard();
  const [problems, setProblems] = useState<Problem[]>([]);
  const [failure, setFailure] = useState<string | null>(null);
  const [hasConflict, setHasConflict] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [leaveTo, setLeaveTo] = useState<string | null>(null);

  useEffect(() => {
    if (leaveTo !== null) {
      void navigate(leaveTo, { replace: true });
    }
  }, [leaveTo, navigate]);

  async function save() {
    if (isSaving) {
      return;
    }
    const found = validateDraft(state.draft, sources);
    setProblems(found);
    setFailure(null);
    if (found.length > 0) {
      return;
    }
    setIsSaving(true);
    try {
      const { name, description, filters, widgets } = state.draft;
      const body = { name: name.trim(), description, filters, widgets };
      const result = dashboardId
        ? await update.mutateAsync({ dashboardId, body: { ...body, version: state.version } })
        : await create.mutateAsync(body);
      dispatch({ type: 'saved', dashboard: result });
      show({ message: 'Dashboard saved' });
      if (!dashboardId) {
        setLeaveTo(`/builder/${result.id}`);
      }
    } catch (error) {
      showFailure(asApiError(error));
    } finally {
      setIsSaving(false);
    }
  }

  function showFailure(error: ApiError) {
    if (error.kind === 'validation') {
      setProblems(serverProblems(error, state.draft.widgets));
    } else if (error.kind === 'nameTaken') {
      setProblems([{ field: 'name', message: error.message }]);
    } else if (error.kind === 'conflict' && dashboardId) {
      setHasConflict(true);
    } else {
      setFailure(error.message);
    }
  }

  async function afterConflict(keepChanges: boolean) {
    setHasConflict(false);
    if (!dashboardId) {
      return;
    }
    try {
      const fresh = await latest(dashboardId);
      dispatch(
        keepChanges
          ? { type: 'followVersion', version: fresh.version }
          : { type: 'load', dashboard: fresh },
      );
      setProblems([]);
    } catch (error) {
      setFailure(asApiError(error).message);
    }
  }

  return {
    save,
    isSaving,
    isLeaving: leaveTo !== null,
    problems,
    failure,
    hasConflict,
    reload: () => afterConflict(false),
    keepEditing: () => afterConflict(true),
  };
}

function serverProblems(error: ApiError, widgets: readonly Widget[]): Problem[] {
  return error.fields.map(({ field, message }) => {
    const index = /^widgets\[(\d+)\]/.exec(field)?.[1];
    const widgetId = index === undefined ? undefined : widgets[Number(index)]?.id;
    return widgetId ? { field, message, widgetId } : { field, message };
  });
}
