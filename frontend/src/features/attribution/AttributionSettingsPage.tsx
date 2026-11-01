import { Lock } from 'lucide-react';
import { useState } from 'react';
import { asApiError, conflictMessage, isSessionLost } from '@/api/errors';
import { localDate } from '@/format/intl';
import { useFormatters } from '@/format/useFormatters';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { useToast } from '@/ui/Toast/useToast';
import { usePageTitle } from '@/ui/usePageTitle';
import { useAttributionSettings, useSaveAttributionSettings } from './api';
import type { AttributionModel, AttributionSettings } from './api';
import styles from './AttributionSettingsPage.module.css';

export const ATTRIBUTION = 'Attribution';
export const ATTRIBUTION_DESCRIPTION =
  'Choose how converted revenue is credited to the activities a customer met before converting.';

/** Settings › Attribution (feature 050): the designated model every attributed figure uses. */
export function AttributionSettingsPage() {
  usePageTitle(ATTRIBUTION);
  const settings = useAttributionSettings();
  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Settings', to: '/settings' }, { label: ATTRIBUTION }]}
        title={ATTRIBUTION}
        description={ATTRIBUTION_DESCRIPTION}
      />
      <Card>
        {settings.data ? (
          <ModelForm settings={settings.data} />
        ) : settings.isError ? (
          <ErrorState error={settings.error} onRetry={() => void settings.refetch()} />
        ) : (
          <Skeleton lines={5} />
        )}
      </Card>
    </>
  );
}

function ModelForm({ settings }: { settings: AttributionSettings }) {
  const save = useSaveAttributionSettings();
  const { show } = useToast();
  const { date, dateFormat } = useFormatters();
  const [model, setModel] = useState<AttributionModel>(settings.model);
  const [confirming, setConfirming] = useState(false);
  const [problem, setProblem] = useState('');
  const locked = !settings.canChange.available;
  const chosen = settings.models.find((m) => m.key === model)?.name ?? model;

  async function send() {
    setProblem('');
    try {
      await save.mutateAsync({ version: settings.version, model });
      show({ message: `Revenue is now credited with ${chosen}` });
    } catch (error) {
      const failure = asApiError(error);
      if (!isSessionLost(failure)) {
        setProblem(conflictMessage(failure, 'setting', dateFormat));
      }
    } finally {
      setConfirming(false);
    }
  }

  return (
    <form
      noValidate
      className={styles.stack}
      onSubmit={(event) => {
        event.preventDefault();
        if (model !== settings.model) {
          setConfirming(true);
        }
      }}
    >
      {problem && (
        <p role="alert" className={styles.problem}>
          {problem}
        </p>
      )}
      <fieldset className={styles.group}>
        <legend className={styles.legend}>Designated model</legend>
        {settings.models.map((m) => (
          <label key={m.key} className={styles.radio}>
            <input
              type="radio"
              name="model"
              value={m.key}
              checked={model === m.key}
              disabled={locked}
              onChange={() => setModel(m.key)}
            />
            <span>
              <span className={styles.word}>{m.name}</span>
              <span className={styles.note}>{m.description}</span>
            </span>
          </label>
        ))}
      </fieldset>
      {locked && (
        <p className={styles.lock} role="note">
          <Lock aria-hidden="true" />
          <span>{settings.canChange.reason}</span>
        </p>
      )}
      <p className={styles.note}>
        {settings.updatedBy && settings.updatedAt
          ? `Last changed by ${settings.updatedBy.name} on ${date(localDate(settings.updatedAt))}.`
          : 'Never changed: Equally distributed is the model MAPP starts with.'}
      </p>
      {!locked && (
        <div className={styles.actions}>
          <Button type="submit" variant="primary" disabled={model === settings.model}>
            Save
          </Button>
          {model === settings.model && (
            <span className={styles.note}>Choose another model to save.</span>
          )}
        </div>
      )}
      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={`Credit revenue with ${chosen}?`}
        confirmLabel="Change the model"
        onConfirm={() => void send()}
        busy={save.isPending}
      >
        Every attributed revenue and ROI figure, on screen and in exports, will be restated with{' '}
        {chosen}.
      </ConfirmDialog>
    </form>
  );
}
