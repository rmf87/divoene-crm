// Inline undo/redo — replaces react-undo-redo (unmaintained, React 19 peer-dep issue).
// Snapshot-based: saves full state before each tracked action.
import {
  useState,
  createContext,
  useContext,
  createElement,
  useCallback,
  useRef,
  type ReactNode,
} from 'react';

type TrackFn<A> = (action: A) => boolean;

interface UndoRedoFn {
  (): void;
  isPossible: boolean;
}

export function createUndoRedo<S, A>(
  reducer: (state: S, action: A) => S,
  options: { track: TrackFn<A> },
) {
  // Contexts created per reducer identity (cached via WeakMap so hot-module reload
  // doesn't re-create them). We use intermediate holder objects to keep references stable.
  const StateCtx = createContext<S | null>(null);
  const DispatchCtx = createContext<((action: A) => void) | null>(null);
  const UndoRedoCtx = createContext<{
    undo: () => void;
    redo: () => void;
    canUndo: boolean;
    canRedo: boolean;
  } | null>(null);

  // ── Provider ──────────────────────────────────────────────────────────────

  function UndoRedoProvider({
    initialState,
    children,
  }: {
    initialState: S;
    children: ReactNode;
  }) {
    const pastRef = useRef<S[]>([]);
    const futureRef = useRef<S[]>([]);
    const [present, setPresent] = useState<S>(initialState);
    const [canUndo, setCanUndo] = useState(false);
    const [canRedo, setCanRedo] = useState(false);

    const dispatch = useCallback(
      (action: A) => {
        setPresent((prev) => {
          const next = reducer(prev, action);
          if (options.track(action)) {
            pastRef.current = [...pastRef.current, prev];
            futureRef.current = [];
            setCanUndo(true);
            setCanRedo(false);
          }
          return next;
        });
      },
      [],
    );

    const undo = useCallback(() => {
      const past = pastRef.current;
      if (past.length === 0) return;
      const previous = past[past.length - 1];
      pastRef.current = past.slice(0, -1);
      setPresent((prev) => {
        futureRef.current = [prev, ...futureRef.current];
        setCanUndo(pastRef.current.length > 0);
        setCanRedo(true);
        return previous;
      });
    }, []);

    const redo = useCallback(() => {
      const future = futureRef.current;
      if (future.length === 0) return;
      const next = future[0];
      futureRef.current = future.slice(1);
      setPresent((prev) => {
        pastRef.current = [...pastRef.current, prev];
        setCanRedo(futureRef.current.length > 0);
        setCanUndo(true);
        return next;
      });
    }, []);

    return createElement(
      StateCtx.Provider,
      { value: present },
      createElement(
        DispatchCtx.Provider,
        { value: dispatch },
        createElement(UndoRedoCtx.Provider, {
          value: { undo, redo, canUndo, canRedo },
        }, children),
      ),
    );
  }

  // ── usePresent ────────────────────────────────────────────────────────────

  function usePresent(): [S, (action: A) => void] {
    const state = useContext(StateCtx);
    const dispatch = useContext(DispatchCtx);
    if (state === null || dispatch === null) {
      throw new Error('usePresent must be used inside UndoRedoProvider');
    }
    return [state, dispatch];
  }

  // ── useUndoRedo ───────────────────────────────────────────────────────────

  function useUndoRedo(): [UndoRedoFn, UndoRedoFn] {
    const ctx = useContext(UndoRedoCtx);
    if (!ctx) {
      throw new Error('useUndoRedo must be used inside UndoRedoProvider');
    }

    // Return callable functions with a reactive .isPossible property.
    // We mutate the same function objects across renders so they're stable references.
    const undoFn: UndoRedoFn = Object.assign(ctx.undo, { isPossible: ctx.canUndo });
    const redoFn: UndoRedoFn = Object.assign(ctx.redo, { isPossible: ctx.canRedo });

    return [undoFn, redoFn];
  }

  return { UndoRedoProvider, usePresent, useUndoRedo };
}
