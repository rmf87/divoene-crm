import { useEffect } from 'react';
import { ArrowClockwise, ArrowCounterClockwise } from '@phosphor-icons/react';

interface UndoRedoControlsProps {
  undo: () => void;
  redo: () => void;
  canUndo: boolean;
  canRedo: boolean;
}

export default function UndoRedoControls({ undo, redo, canUndo, canRedo }: UndoRedoControlsProps) {
  // Keyboard shortcuts: Ctrl+Z / Cmd+Z for undo, Ctrl+Shift+Z for redo
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'z') {
        if (e.shiftKey) {
          e.preventDefault();
          if (canRedo) redo();
        } else {
          e.preventDefault();
          if (canUndo) undo();
        }
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [undo, redo, canUndo, canRedo]);

  return (
    <div className="flex items-center gap-1" id="undo-redo-controls">
      <button
        id="btn-undo"
        onClick={undo}
        disabled={!canUndo}
        title="Desfazer (Ctrl+Z)"
        className="px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-1.5
          enabled:bg-earth-100 enabled:text-olive-700 enabled:hover:bg-earth-200
          disabled:opacity-40 disabled:cursor-not-allowed"
      >
        <ArrowCounterClockwise size={16} weight="bold" />
        Desfazer
      </button>
      <button
        id="btn-redo"
        onClick={redo}
        disabled={!canRedo}
        title="Refazer (Ctrl+Shift+Z)"
        className="px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-1.5
          enabled:bg-earth-100 enabled:text-olive-700 enabled:hover:bg-earth-200
          disabled:opacity-40 disabled:cursor-not-allowed"
      >
        <ArrowClockwise size={16} weight="bold" />
        Refazer
      </button>
    </div>
  );
}
