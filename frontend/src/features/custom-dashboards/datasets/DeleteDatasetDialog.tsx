import { ApiError, isSessionLost } from '@/api/errors';
import { formatWhole } from '@/format/decimal';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { useToast } from '@/ui/Toast/useToast';
import { useUndoableDelete } from '@/ui/useUndoableDelete';
import { dataSourceKeys, datasetKeys } from '../keys';
import type { Dataset } from '../model';
import { useDeleteDataset } from './api';

interface DeleteDatasetDialogProps {
  dataset: Dataset;
  onClose: () => void;
  onDeleted: () => void;
}

export function DeleteDatasetDialog({ dataset, onClose, onDeleted }: DeleteDatasetDialogProps) {
  const remove = useDeleteDataset();
  const { show } = useToast();
  const afterDelete = useUndoableDelete({ invalidate: [datasetKeys.all, dataSourceKeys.all] });
  const used = dataset.usedBy.length;

  async function confirm() {
    if (remove.isPending) {
      return;
    }
    try {
      afterDelete(await remove.mutateAsync(dataset.id));
      onClose();
      onDeleted();
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
      title={`Delete dataset '${dataset.name}'?`}
      confirmLabel="Delete dataset"
      tone="danger"
      busy={remove.isPending}
      onConfirm={() => void confirm()}
    >
      {used > 0 &&
        `${usageText(used)}; their widgets will show 'This data is no longer available'. `}
      It moves to the recycle bin; you can restore it.
    </ConfirmDialog>
  );
}

function usageText(count: number): string {
  return count === 1
    ? '1 of your dashboards uses it'
    : `${formatWhole(count)} of your dashboards use it`;
}
