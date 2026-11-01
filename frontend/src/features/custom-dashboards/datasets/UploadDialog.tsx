import { CircleAlert } from 'lucide-react';
import { useState } from 'react';
import { ApiError, isSessionLost } from '@/api/errors';
import type { FieldProblem } from '@/ui/FormErrorSummary/FormErrorSummary';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import type { Dataset, DatasetPreview } from '../model';
import { usePreviewDataset } from './api';
import type { DatasetUpload } from './api';
import { ColumnsStep } from './ColumnsStep';
import { base64Of, fileProblem, LIMITS_TEXT } from './files';
import styles from './datasets.module.css';

const FILE_FIELD = 'dataset-file';

interface UploadDialogProps {
  onClose: () => void;
  onSaved: (dataset: Dataset) => void;
}

interface Checked {
  upload: DatasetUpload;
  preview: DatasetPreview;
}

export function UploadDialog({ onClose, onSaved }: UploadDialogProps) {
  const [checked, setChecked] = useState<Checked | null>(null);
  if (checked) {
    return (
      <ColumnsStep
        upload={checked.upload}
        preview={checked.preview}
        onBack={() => setChecked(null)}
        onClose={onClose}
        onSaved={onSaved}
      />
    );
  }
  return <ChooseFileStep onChecked={setChecked} onClose={onClose} />;
}

function ChooseFileStep({
  onChecked,
  onClose,
}: {
  onChecked: (checked: Checked) => void;
  onClose: () => void;
}) {
  const preview = usePreviewDataset();
  const [file, setFile] = useState<File | null>(null);
  const [problems, setProblems] = useState<FieldProblem[]>([]);
  const [failure, setFailure] = useState<SaveFailure | null>(null);

  async function check() {
    setFailure(null);
    const problem = fileProblem(file);
    if (problem || !file) {
      setProblems([{ field: FILE_FIELD, message: problem ?? '' }]);
      return;
    }
    setProblems([]);
    try {
      const upload = { fileName: file.name, content: await base64Of(file) };
      onChecked({ upload, preview: await preview.mutateAsync(upload) });
    } catch (error) {
      showRefusal(error instanceof ApiError ? error : new ApiError('other'));
    }
  }

  function showRefusal(error: ApiError) {
    if (error.kind === 'validation' || error.kind === 'tooLarge') {
      const messages =
        error.fields.length > 0 ? error.fields.map((f) => f.message) : [error.message];
      setProblems(messages.map((message) => ({ field: FILE_FIELD, message })));
    } else if (!isSessionLost(error)) {
      setFailure({ message: error.message });
    }
  }

  return (
    <FormPanel
      title="Upload data"
      description="Step 1 of 2: Choose a file"
      isDirty={false}
      saving={preview.isPending}
      problems={problems}
      failure={failure}
      onSave={() => void check()}
      onClose={onClose}
      saveLabel="Check the file"
    >
      <div className={styles.fileField}>
        <label htmlFor={FILE_FIELD} className={styles.legend}>
          Choose a file
        </label>
        <p id={`${FILE_FIELD}-hint`} className={styles.facts}>
          {LIMITS_TEXT} SP&amp;A One reads the first sheet; the first row names the columns.
        </p>
        <input
          id={FILE_FIELD}
          type="file"
          accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
          aria-describedby={`${FILE_FIELD}-hint`}
          aria-invalid={problems.length > 0 || undefined}
          onChange={(event) => setFile(event.target.files?.[0] ?? null)}
        />
        {problems.map((problem) => (
          <p key={problem.message} className={styles.problem}>
            <CircleAlert aria-hidden="true" />
            {problem.message}
          </p>
        ))}
      </div>
    </FormPanel>
  );
}
