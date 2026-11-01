import { useState } from 'react';
import { asApiError } from '@/api/errors';
import { FormDialog } from '@/ui/FormDialog/FormDialog';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import styles from '../activity/DataSourcesArea.module.css';
import { useGenerateMediaCode, usePerformanceConfiguration } from './api';

/** Generates a code with the configuration's attributes; every missing one is reported at once (FR-041). */
export function GenerateCodeDialog({
  activityId,
  onClose,
}: {
  activityId: string;
  onClose: () => void;
}) {
  const { show } = useToast();
  const configuration = usePerformanceConfiguration();
  const generate = useGenerateMediaCode(activityId);
  const [values, setValues] = useState<Record<string, string>>({});
  const [landingPage, setLandingPage] = useState('');
  const [problems, setProblems] = useState<Record<string, string>>({});

  function save() {
    generate.mutate(
      { attributes: values, landingPage },
      {
        onSuccess: (code) => {
          show({ message: `Code ${code.code} generated` });
          onClose();
        },
        onError: (error) => {
          const failed = asApiError(error);
          const next: Record<string, string> = {};
          for (const field of failed.fields) {
            next[field.field.replace('attributes.', '')] = field.message;
          }
          setProblems(next);
          if (failed.fields.length === 0) {
            show({ message: failed.message });
          }
        },
      },
    );
  }

  const attributes = configuration.data?.codeAttributes ?? [];
  return (
    <FormDialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title="Generate media code"
      submitLabel="Generate"
      busy={generate.isPending}
      onSubmit={save}
    >
      <div className={styles.fields}>
        {attributes.map((attribute) => (
          <TextField
            key={attribute.name}
            label={attribute.name}
            name={`attribute-${attribute.name}`}
            marking={attribute.required ? 'Required' : undefined}
            value={values[attribute.name] ?? ''}
            onChange={(value) => setValues((held) => ({ ...held, [attribute.name]: value }))}
            error={problems[attribute.name]}
          />
        ))}
        <TextField
          label="Landing page"
          name="landingPage"
          hint="The code is appended to this address as the tagged link."
          value={landingPage}
          onChange={setLandingPage}
        />
      </div>
    </FormDialog>
  );
}
