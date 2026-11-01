import { zodResolver } from '@hookform/resolvers/zod';
import { useId, useState } from 'react';
import { useForm } from 'react-hook-form';
import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { ConfirmDialog } from '@/ui/ConfirmDialog/ConfirmDialog';
import { FormDialog } from '@/ui/FormDialog/FormDialog';
import { FormErrorSummary } from '@/ui/FormErrorSummary/FormErrorSummary';
import { FormField } from '@/ui/FormField/FormField';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { ReportKey } from '../reports';
import { useCreateReportView, useUpdateReportView } from './api';
import type { ReportView } from './api';
import { saveViewSchema, useViewRefusal, viewNamed } from './saved-views';
import type { SaveViewValues } from './saved-views';

const FIELDS = ['name'] as const;

interface SaveViewDialogProps {
  report: ReportKey;
  views: ReportView[];
  /** The address on screen, which the view keeps. */
  query: string;
  onOpenChange: (open: boolean) => void;
  onSaved: (view: ReportView) => void;
}

/** Save current filters (FR-024, FR-027): a name in use is replaced only after asking. Mounted while open. */
export function SaveViewDialog({
  report,
  views,
  query,
  onOpenChange,
  onSaved,
}: SaveViewDialogProps) {
  const id = useId();
  const nameField = `${id}-name`;
  const form = useForm<SaveViewValues>({
    resolver: zodResolver(saveViewSchema),
    defaultValues: { name: '', isDefault: false },
    ...CHECK_ON_SAVE,
  });
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, FIELDS);
  const [replacing, setReplacing] = useState<ReportView | null>(null);
  const create = useCreateReportView(report);
  const update = useUpdateReportView(report);
  const refused = useViewRefusal();
  const { show } = useToast();

  function done(view: ReportView, verb: 'saved' | 'updated') {
    show({ message: `View '${view.name}' ${verb}` });
    onSaved(view);
    onOpenChange(false);
  }

  const submit = form.handleSubmit(({ name, isDefault }) => {
    clearProblems();
    const existing = viewNamed(views, name);
    if (existing) {
      setReplacing(existing);
      return;
    }
    create.mutate(
      { name: name.trim(), query, isDefault },
      {
        onSuccess: (view) => done(view, 'saved'),
        onError: (error) => showServerProblems(refused(error, 'name')),
      },
    );
  }, onInvalid);

  function replace(view: ReportView) {
    const isDefault = form.getValues('isDefault');
    update.mutate(
      { viewId: view.id, body: { version: view.version, query, ...(isDefault && { isDefault }) } },
      {
        onSuccess: (saved) => done(saved, 'updated'),
        onError: (error) => {
          setReplacing(null);
          showServerProblems(refused(error, 'name'));
        },
      },
    );
  }

  return (
    <>
      <FormDialog
        open
        onOpenChange={onOpenChange}
        title="Save current filters"
        submitLabel="Save view"
        onSubmit={() => void submit()}
        busy={create.isPending}
      >
        <FormErrorSummary
          errors={problems}
          fieldId={(field) => (field === 'name' ? nameField : field)}
        />
        <FormField
          control={form.control}
          name="name"
          render={(field) => (
            <TextField label="Name" marking="Required" {...field} name={nameField} />
          )}
        />
        <FormField
          control={form.control}
          name="isDefault"
          render={(field) => (
            <CheckboxField
              label="Open this report on this view"
              name={`${id}-default`}
              checked={field.value}
              onChange={field.onChange}
            />
          )}
        />
      </FormDialog>
      {replacing && (
        <ConfirmDialog
          open
          onOpenChange={(open) => !open && setReplacing(null)}
          title={`Replace your view '${replacing.name}'?`}
          confirmLabel="Replace view"
          onConfirm={() => replace(replacing)}
          busy={update.isPending}
        >
          It will keep the filters you see now.
        </ConfirmDialog>
      )}
    </>
  );
}
