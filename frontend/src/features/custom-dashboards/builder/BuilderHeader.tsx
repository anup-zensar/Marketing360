import { Eye, Redo2, Share2, Undo2 } from 'lucide-react';
import { Button, LinkButton } from '@/ui/Button/Button';
import { PageHeader } from '@/ui/PageHeader/PageHeader';

export interface BuilderHeaderProps {
  title: string;
  dashboardId: string | undefined;
  canShare: boolean;
  isSaving: boolean;
  onUndo: () => void;
  onRedo: () => void;
  onSharing: () => void;
  onSave: () => void;
}

export function BuilderHeader({
  title,
  dashboardId,
  canShare,
  isSaving,
  onUndo,
  onRedo,
  onSharing,
  onSave,
}: BuilderHeaderProps) {
  return (
    <PageHeader
      breadcrumb={[{ label: 'Dashboards', to: '/dashboards' }, { label: title }]}
      title={title}
      actions={
        <>
          <Button icon={<Undo2 />} aria-keyshortcuts="Control+Z Meta+Z" onClick={onUndo}>
            Undo
          </Button>
          <Button
            icon={<Redo2 />}
            aria-keyshortcuts="Control+Shift+Z Meta+Shift+Z Control+Y"
            onClick={onRedo}
          >
            Redo
          </Button>
          {dashboardId && canShare && (
            <Button icon={<Share2 />} onClick={onSharing}>
              Sharing
            </Button>
          )}
          {dashboardId && (
            <LinkButton icon={<Eye />} to={`/dashboards/${dashboardId}`}>
              View
            </LinkButton>
          )}
          <Button variant="primary" aria-busy={isSaving} onClick={onSave}>
            {isSaving ? 'Saving…' : 'Save'}
          </Button>
        </>
      }
    />
  );
}
