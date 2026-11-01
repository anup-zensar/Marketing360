import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import styles from './ApprovalsGrid.module.css';

export interface CommentDialogProps {
  title: string | null;
  comment: string;
  onClose: () => void;
}

/** A long comment in full, on the page; Escape or Close returns to the grid. */
export function CommentDialog({ title, comment, onClose }: CommentDialogProps) {
  return (
    <ConfirmDialog
      open={title !== null}
      onOpenChange={(open) => !open && onClose()}
      title={title ?? ''}
      acknowledgeOnly
    >
      <p className={styles.comment}>{comment}</p>
    </ConfirmDialog>
  );
}
