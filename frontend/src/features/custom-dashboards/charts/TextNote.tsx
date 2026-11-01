import styles from './WidgetView.module.css';

interface TextNoteProps {
  note: string;
  isEditing: boolean;
}

/** Plain text only: React escapes it, and line breaks are kept by CSS, never turned into HTML. */
export function TextNote({ note, isEditing }: TextNoteProps) {
  if (note.trim() === '') {
    return isEditing ? (
      <p className={styles.placeholder}>Write the note in the Style tab.</p>
    ) : null;
  }
  return <p className={styles.textNote}>{note}</p>;
}
