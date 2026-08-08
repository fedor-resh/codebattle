import { MantineProvider } from '@mantine/core';
import { act, render, screen, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

const editorMock = vi.hoisted(() => {
  const state: {
    cursorListener?: (event: { position: { lineNumber: number; column: number } }) => void;
  } = {};
  return {
    state,
    addContentWidget: vi.fn(),
    removeContentWidget: vi.fn(),
    editor: {
      getModel: () => ({
        getLineCount: () => 20,
        getLineMaxColumn: () => 100,
      }),
      onDidChangeCursorPosition: vi.fn(
        (listener: (event: { position: { lineNumber: number; column: number } }) => void) => {
          state.cursorListener = listener;
          return { dispose: vi.fn() };
        },
      ),
      addContentWidget: (...args: unknown[]) => editorMock.addContentWidget(...args),
      removeContentWidget: (...args: unknown[]) => editorMock.removeContentWidget(...args),
    },
  };
});

vi.mock('../monaco', () => ({}));
vi.mock('@monaco-editor/react', async () => {
  const React = await import('react');
  return {
    default: ({ onMount }: { onMount?: (editor: unknown) => void }) => {
      React.useEffect(() => onMount?.(editorMock.editor), [onMount]);
      return React.createElement('div', { 'data-testid': 'monaco-editor' });
    },
  };
});

import { CodePane } from './CodePane';

describe('CodePane', () => {
  it('shows the opponent cursor and reports local cursor movement', async () => {
    const onCursorChange = vi.fn();
    render(
      <MantineProvider>
        <CodePane
          label="Код соперника"
          value={'package solution\n\nfunc Solve() {}'}
          readOnly
          onCursorChange={onCursorChange}
          remoteCursor={{ lineNumber: 4, column: 7, label: 'rival' }}
        />
      </MantineProvider>,
    );

    expect(screen.getByText('Курсор 4:7')).toBeInTheDocument();
    await waitFor(() => expect(editorMock.addContentWidget).toHaveBeenCalledOnce());

    const widget = editorMock.addContentWidget.mock.calls[0][0] as {
      getDomNode: () => HTMLElement;
      getPosition: () => { position: { lineNumber: number; column: number } };
    };
    expect(widget.getDomNode()).toHaveTextContent('rival');
    expect(widget.getPosition().position).toEqual({ lineNumber: 4, column: 7 });

    act(() => editorMock.state.cursorListener?.({ position: { lineNumber: 6, column: 3 } }));
    expect(onCursorChange).toHaveBeenCalledWith({ lineNumber: 6, column: 3 });
  });
});
