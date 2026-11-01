import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { DecimalField } from '@/ui/DecimalField/DecimalField';
import { FormField } from '@/ui/FormField/FormField';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import { MoneyField } from '@/ui/MoneyField/MoneyField';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import { parseDecimalInput, parseMoneyInput } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import styles from '@/ui/settings-table.module.css';
import { ImpactDialog } from '../ImpactDialog';
import { saveFailureOf } from '../save-failure';
import { ActivityTypesField } from './ActivityTypesField';
import { GoalStagesField } from './GoalStagesField';
import { checkModelImpact, fetchModel, useCreateModel, useUpdateModel } from './api';
import type { ChangeImpact, Model, ModelUpdate } from './api';
import { affectsOthers, changedFields, createBody, startValues } from './model-values';
import type { ModelFormMode } from './model-values';
import { modelSchema } from './schemas';
import type { ModelValues } from './schemas';
import { WaterfallExample } from './WaterfallExample';

const FIELDS = [
  'name',
  'inqToMqlRate',
  'mqlToDealRate',
  'averageDealSize',
  'activityTypeIds',
  'inqToMqlVelocityMonths',
  'mqlToDealVelocityMonths',
  'allowedGoalStages',
] as const;

export interface ModelFormProps {
  mode: ModelFormMode;
  /** Opens a model's details: the View of the "created" message. */
  onView: (id: string) => void;
  onClose: () => void;
}

/** A save that waits for "Save changes" in the impact dialog. */
interface PendingImpact {
  body: ModelUpdate;
  warnings: ChangeImpact['warnings'];
}

/**
 * New model, Edit model and Duplicate (FR-009 – FR-012) in a side panel: name, both rates, the
 * average deal size, the allowed activity types and the live worked example. Every problem shows at
 * once; a change that affects plans or activities is confirmed first; a conflict keeps what was
 * typed and names who saved first.
 */
/** A typed velocity, or null while it is not a whole number — the example then shows no lag. */
function wholeMonths(typed: string): number | null {
  const months = Number.parseInt(typed.trim(), 10);
  return Number.isInteger(months) && months >= 0 ? months : null;
}

export function ModelForm({ mode, onView, onClose }: ModelFormProps) {
  const { show } = useToast();
  const { dateFormat } = useFormatters();
  const createModel = useCreateModel();
  const updateModel = useUpdateModel();
  const editing = mode.kind === 'edit' ? mode.model : null;
  // The edit-form rule: the model is copied into the form once, when the panel opens.
  const [start] = useState(() => startValues(mode));
  const form = useForm<ModelValues>({
    resolver: zodResolver(modelSchema),
    defaultValues: start,
    ...CHECK_ON_SAVE,
  });
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, FIELDS);
  const [version, setVersion] = useState(editing?.version ?? 0);
  const [saving, setSaving] = useState(false);
  const [failure, setFailure] = useState<SaveFailure | null>(null);
  const [pendingImpact, setPendingImpact] = useState<PendingImpact | null>(null);
  const [inqToMqlRate, mqlToDealRate, averageDealSize, inqVelocity, mqlVelocity] = useWatch({
    control: form.control,
    name: [
      'inqToMqlRate',
      'mqlToDealRate',
      'averageDealSize',
      'inqToMqlVelocityMonths',
      'mqlToDealVelocityMonths',
    ],
  });

  async function save(values: ModelValues) {
    clearProblems();
    setSaving(true);
    try {
      if (!editing) {
        const created = await createModel.mutateAsync(createBody(values));
        show({
          message: `Model '${created.name}' created`,
          action: {
            label: 'View',
            altText: 'Find it in the list of models',
            onClick: () => onView(created.id),
          },
        });
        onClose();
        return;
      }
      const body = changedFields(editing, start, values, version);
      if (!body) {
        onClose();
        return;
      }
      if (affectsOthers(body)) {
        const impact = await checkModelImpact(editing.id, body);
        if (impact.warnings.length > 0) {
          setPendingImpact({ body, warnings: impact.warnings });
          return;
        }
      }
      await saveChanges(editing, body);
    } catch (error) {
      await showFailure(error);
    } finally {
      setSaving(false);
    }
  }

  async function saveChanges(model: Model, body: ModelUpdate) {
    const saved = await updateModel.mutateAsync({ id: model.id, body });
    show({ message: `Model '${saved.name}' saved` });
    onClose();
  }

  async function confirmImpact() {
    if (!editing || !pendingImpact) {
      return;
    }
    setSaving(true);
    try {
      await saveChanges(editing, pendingImpact.body);
    } catch (error) {
      await showFailure(error);
    } finally {
      setSaving(false);
      setPendingImpact(null);
    }
  }

  async function showFailure(error: unknown) {
    const view = saveFailureOf(error, 'model', dateFormat);
    showServerProblems(view.problems);
    setFailure(view.failure);
    if (view.conflict && editing) {
      // Typed values stay; the next Save replaces the other person's change on purpose.
      const latest = await fetchModel(editing.id).catch(() => undefined);
      if (latest) {
        setVersion(latest.version);
      }
    }
  }

  const submit = form.handleSubmit(save, onInvalid);

  return (
    <>
      <FormPanel
        title={editing ? 'Edit model' : 'New model'}
        isDirty={form.formState.isDirty}
        saving={saving}
        problems={problems}
        failure={failure}
        onSave={() => void submit()}
        onClose={onClose}
      >
        <div className={styles.formFields}>
          <FormField
            control={form.control}
            name="name"
            render={(field) => <TextField label="Name" marking="Required" {...field} />}
          />
          <FormField
            control={form.control}
            name="inqToMqlRate"
            render={(field) => (
              <DecimalField
                label="INQ→MQL rate"
                marking="Required"
                decimals={2}
                suffix="%"
                {...field}
              />
            )}
          />
          <FormField
            control={form.control}
            name="mqlToDealRate"
            render={(field) => (
              <DecimalField
                label="MQL→deal rate"
                marking="Required"
                decimals={2}
                suffix="%"
                {...field}
              />
            )}
          />
          <FormField
            control={form.control}
            name="averageDealSize"
            render={(field) => (
              <MoneyField label="Average deal size" marking="Required" currency="USD" {...field} />
            )}
          />
          <FormField
            control={form.control}
            name="inqToMqlVelocityMonths"
            render={(field) => (
              <DecimalField
                label="INQ→MQL velocity"
                decimals={0}
                suffix="months"
                note="How long a conversion takes, so a projection lands in the right month."
                {...field}
              />
            )}
          />
          <FormField
            control={form.control}
            name="mqlToDealVelocityMonths"
            render={(field) => (
              <DecimalField label="MQL→deal velocity" decimals={0} suffix="months" {...field} />
            )}
          />
          <FormField
            control={form.control}
            name="allowedGoalStages"
            render={(field) => <GoalStagesField {...field} />}
          />
          <FormField
            control={form.control}
            name="activityTypeIds"
            render={(field) => <ActivityTypesField model={editing} {...field} />}
          />
          <WaterfallExample
            inqToMqlRate={parseDecimalInput(inqToMqlRate, 2)}
            mqlToDealRate={parseDecimalInput(mqlToDealRate, 2)}
            averageDealSize={parseMoneyInput(averageDealSize)?.amount ?? null}
            inqToMqlVelocityMonths={wholeMonths(inqVelocity)}
            mqlToDealVelocityMonths={wholeMonths(mqlVelocity)}
          />
        </div>
      </FormPanel>
      {pendingImpact && editing && (
        <ImpactDialog
          kind="model"
          name={editing.name}
          warnings={pendingImpact.warnings}
          busy={saving}
          onConfirm={() => void confirmImpact()}
          onCancel={() => setPendingImpact(null)}
        />
      )}
    </>
  );
}
