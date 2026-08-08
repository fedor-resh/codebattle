import { MantineProvider } from '@mantine/core';
import { act, render, screen, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

const autocompleteMock = vi.hoisted(() => ({
  dispose: vi.fn(),
  register: vi.fn(() => ({ dispose: autocompleteMock.dispose })),
}));
const editorMock = vi.hoisted(() => {
  const state: {
    cursorListener?: (event: { position: { lineNumber: number; column: number } }) => void;
    editorOnChange?: (value?: string) => void;
    value: string;
  } = { value: '' };
  return {
    state,
    addContentWidget: vi.fn(),
    removeContentWidget: vi.fn(),
    setModelValue: vi.fn((value: string) => {
      state.value = value;
      state.editorOnChange?.(value);
    }),
    editor: {
      getModel: () => ({
        getLineCount: () => 20,
        getLineMaxColumn: () => 100,
        getValue: () => state.value,
        setValue: (value: string) => editorMock.setModelValue(value),
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
vi.mock('../goAutocomplete', () => ({
  registerGoAutocomplete: autocompleteMock.register,
}));
vi.mock('@monaco-editor/react', async () => {
  const React = await import('react');
  return {
    default: ({
      defaultValue,
      onChange,
      onMount,
    }: {
      defaultValue?: string;
      onChange?: (value?: string) => void;
      onMount?: (editor: unknown) => void;
    }) => {
      const initialValue = React.useRef(defaultValue ?? '');
      React.useEffect(() => {
        editorMock.state.value = initialValue.current;
        onMount?.(editorMock.editor);
      }, [onMount]);
      React.useEffect(() => {
        editorMock.state.editorOnChange = onChange;
      }, [onChange]);
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
    expect(autocompleteMock.register).not.toHaveBeenCalled();
  });

  it('registers autocomplete only while an editable editor is mounted', async () => {
    const { unmount } = render(
      <MantineProvider>
        <CodePane
          label="Ваш код"
          value={'package solution\n\nfunc Solve() {}'}
          functionSignature="func Solve(input string) string"
        />
      </MantineProvider>,
    );

    await waitFor(() => expect(autocompleteMock.register).toHaveBeenCalledOnce());
    expect(autocompleteMock.register).toHaveBeenCalledWith(
      expect.anything(),
      editorMock.editor,
      'func Solve(input string) string',
    );

    unmount();
    expect(autocompleteMock.dispose).toHaveBeenCalledOnce();
  });

  it('applies external source updates without echoing them through onChange', async () => {
    const onChange = vi.fn();
    const { rerender } = render(
      <MantineProvider>
        <CodePane label="Ваш код" value="first" onChange={onChange} />
      </MantineProvider>,
    );

    rerender(
      <MantineProvider>
        <CodePane label="Ваш код" value="second" onChange={onChange} />
      </MantineProvider>,
    );

    await waitFor(() => expect(editorMock.setModelValue).toHaveBeenCalledWith('second'));
    expect(onChange).not.toHaveBeenCalled();
  });
});
