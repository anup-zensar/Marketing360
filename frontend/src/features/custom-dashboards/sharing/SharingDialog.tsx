import type { ReactElement } from 'react';
import type { TeamRef } from '../model';

export interface SharingDialogProps {
  open: boolean;
  recordLabel: string;
  recordName: string;
  teams: TeamRef[];
  selected: string[];
  saving: boolean;
  error: string | null;
  onSave: (teamIds: string[]) => void;
  onClose: () => void;
}

export function SharingDialog({ open }: SharingDialogProps): ReactElement {
  return <>{open ? <p>Coming in this feature</p> : null}</>;
}
