import { useEffect } from 'react';
import type { BuilderAction } from './builder-reducer';

const TYPING_TAGS = ['INPUT', 'TEXTAREA', 'SELECT'];

export function useEditorShortcuts(dispatch: (action: BuilderAction) => void): void {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      const command = event.ctrlKey || event.metaKey;
      if (!command || leftToTheBrowser(event.target)) {
        return;
      }
      const key = event.key.toLowerCase();
      const redo = (key === 'z' && event.shiftKey) || (key === 'y' && event.ctrlKey);
      if (key === 'z' || redo) {
        event.preventDefault();
        dispatch({ type: redo ? 'redo' : 'undo' });
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [dispatch]);
}

function leftToTheBrowser(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }
  const typing = target.isContentEditable || TYPING_TAGS.includes(target.tagName);
  return typing || target.closest('[role="dialog"], [role="alertdialog"]') !== null;
}
