import { CircleAlert } from 'lucide-react';
import { useState } from 'react';
import type { ReactElement } from 'react';
import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { FormDialog } from '@/ui/FormDialog/FormDialog';
import type { TeamRef } from '../model';
import styles from './SharingDialog.module.css';

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

export function SharingDialog(props: SharingDialogProps): ReactElement {
  return <>{props.open && <SharingForm {...props} />}</>;
}

function SharingForm({
  recordLabel,
  recordName,
  teams,
  selected,
  saving,
  error,
  onSave,
  onClose,
}: SharingDialogProps) {
  const [chosen, setChosen] = useState<string[]>(selected);

  function toggle(teamId: string, isChecked: boolean) {
    setChosen((current) =>
      isChecked ? [...current, teamId] : current.filter((id) => id !== teamId),
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(isOpen) => !isOpen && onClose()}
      title={`Share '${recordName}'`}
      submitLabel="Save"
      busy={saving}
      onSubmit={() => onSave(teams.filter((team) => chosen.includes(team.id)).map((t) => t.id))}
    >
      {teams.length === 0 ? (
        <p className={styles.note}>
          You aren&apos;t in any team yet, so there is no one to share this {recordLabel} with.
        </p>
      ) : (
        <fieldset className={styles.teams}>
          <legend className={styles.legend}>Teams</legend>
          {teams.map((team) => (
            <CheckboxField
              key={team.id}
              name={`share-${team.id}`}
              label={team.active ? team.name : `${team.name} (inactive)`}
              checked={chosen.includes(team.id)}
              onChange={(isChecked) => toggle(team.id, isChecked)}
            />
          ))}
        </fieldset>
      )}
      <p className={styles.note}>
        People in these teams can view this {recordLabel}. Each of them sees only the data they may
        see.
      </p>
      {error && (
        <p role="alert" className={styles.error}>
          <CircleAlert aria-hidden="true" />
          {error}
        </p>
      )}
    </FormDialog>
  );
}
