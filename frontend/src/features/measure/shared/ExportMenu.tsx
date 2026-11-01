import { Download, LoaderCircle } from 'lucide-react';
import { DropdownMenu } from 'radix-ui';
import { useRef, useState } from 'react';
import { ApiError } from '@/api/errors';
import { refusalText } from '@/session/refusal-text';
import { useAccess } from '@/session/access';
import { Button } from '@/ui/Button/Button';
import { useToast } from '@/ui/Toast/useToast';
import type { ExportFormat } from './api';
import styles from './ReportFrame.module.css';

const FORMAT_WORDS: Record<ExportFormat, string> = { xlsx: 'Excel (.xlsx)', csv: 'CSV (.csv)' };

interface ExportMenuProps {
  label: string;
  formats: ExportFormat[];
  download: (format: ExportFormat) => Promise<string>;
}

/** Export, or the reason it is unavailable: "Exporting isn't part of your role" (FR-022). */
export function ExportMenu({ label, formats, download }: ExportMenuProps) {
  const access = useAccess();
  const { show } = useToast();
  const [preparing, setPreparing] = useState(false);
  const chosen = useRef<ExportFormat | null>(null);

  if (!access.has('data.export')) {
    const reason = refusalText('Exporting', access);
    return (
      <Button
        variant="secondary"
        size="sm"
        icon={<Download />}
        disabled
        title={reason}
        aria-description={reason}
      >
        {label}
      </Button>
    );
  }

  async function exportAs(format: ExportFormat) {
    setPreparing(true);
    try {
      show({ message: `Exported '${await download(format)}'` });
    } catch (error) {
      const failure = error instanceof ApiError ? error : new ApiError('other');
      if (failure.kind !== 'signInRequired' && failure.kind !== 'sessionExpired') {
        show({ message: failure.message });
      }
    } finally {
      setPreparing(false);
    }
  }

  return (
    <DropdownMenu.Root modal={false}>
      <DropdownMenu.Trigger asChild>
        <Button
          variant="secondary"
          size="sm"
          icon={preparing ? <LoaderCircle /> : <Download />}
          aria-busy={preparing || undefined}
        >
          {preparing ? 'Preparing your file…' : label}
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          className={styles.menu}
          align="end"
          sideOffset={4}
          onCloseAutoFocus={() => {
            const format = chosen.current;
            chosen.current = null;
            if (format) {
              void exportAs(format);
            }
          }}
        >
          {formats.map((format) => (
            <DropdownMenu.Item
              key={format}
              className={styles.item}
              disabled={preparing}
              onSelect={() => {
                chosen.current = format;
              }}
            >
              {FORMAT_WORDS[format]}
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
