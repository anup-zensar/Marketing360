import { useState } from 'react';
import { asApiError } from '@/api/errors';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import { FormDialog } from '@/ui/FormDialog/FormDialog';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { usePerformanceSources } from '../api';
import { useSetIntegrationId } from './api';
import styles from './DataSourcesArea.module.css';

/** Registers one source's own id for the activity (FR-037); a second for the same source answers 409. */
export function KeyDialog({ activityId, onClose }: { activityId: string; onClose: () => void }) {
  const { show } = useToast();
  const sources = usePerformanceSources();
  const setKey = useSetIntegrationId(activityId);
  const [sourceId, setSourceId] = useState('');
  const [value, setValue] = useState('');
  const [problem, setProblem] = useState('');

  function save() {
    if (!sourceId || !value.trim()) {
      setProblem('Choose the data source and give its id for this activity.');
      return;
    }
    setKey.mutate(
      { sourceId, value: value.trim() },
      {
        onSuccess: () => {
          show({ message: 'Integration ID registered' });
          onClose();
        },
        onError: (error) => setProblem(asApiError(error).message),
      },
    );
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title="Register integration ID"
      submitLabel="Register"
      busy={setKey.isPending}
      onSubmit={save}
    >
      <div className={styles.fields}>
        <ChoiceField
          label="Data source"
          name="sourceId"
          options={(sources.data?.items ?? []).map((s) => ({
            id: s.id,
            label: s.name,
            active: true,
          }))}
          value={sourceId}
          onChange={setSourceId}
        />
        <TextField
          label="Integration ID"
          name="integrationValue"
          hint="The source's own id for this activity, exactly as its rows will carry it."
          value={value}
          onChange={setValue}
          error={problem || undefined}
        />
      </div>
    </FormDialog>
  );
}
