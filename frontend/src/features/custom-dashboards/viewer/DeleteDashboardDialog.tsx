import { ApiError, isSessionLost } from '@/api/errors';
import { joinWithAnd } from '@/format/text';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { useToast } from '@/ui/Toast/useToast';
import { useUndoableDelete } from '@/ui/useUndoableDelete';
import { dashboardKeys } from '../keys';
import type { TeamRef } from '../model';
import { useDeleteCustomDashboard } from './api';

export interface DashboardToDelete {
  id: string;
  name: string;
  sharedWith: TeamRef[];
}

interface DeleteDashboardDialogProps {
  dashboard: DashboardToDelete;
  onClose: () => void;
  onDeleted?: () => void;
}

export function DeleteDashboardDialog({
  dashboard,
  onClose,
  onDeleted,
}: DeleteDashboardDialogProps) {
  const remove = useDeleteCustomDashboard();
  const { show } = useToast();
  const afterDelete = useUndoableDelete({ invalidate: [dashboardKeys.all] });
  const teams = joinWithAnd(dashboard.sharedWith.map((team) => team.name));

  async function confirm() {
    if (remove.isPending) {
      return;
    }
    try {
      afterDelete(await remove.mutateAsync(dashboard.id));
      onClose();
      onDeleted?.();
    } catch (error) {
      const failure = error instanceof ApiError ? error : new ApiError('other');
      if (!isSessionLost(failure)) {
        show({ message: failure.message });
      }
      onClose();
    }
  }

  return (
    <ConfirmDialog
      open
      onOpenChange={(isOpen) => !isOpen && onClose()}
      title={`Delete dashboard '${dashboard.name}'?`}
      confirmLabel="Delete dashboard"
      tone="danger"
      busy={remove.isPending}
      onConfirm={() => void confirm()}
    >
      {teams ? `It's shared with ${teams}; they will no longer see it. ` : ''}
      It moves to the recycle bin; you can restore it.
    </ConfirmDialog>
  );
}
