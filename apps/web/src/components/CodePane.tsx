import Editor from '@monaco-editor/react';
import { Badge, Group, Paper, Text, useComputedColorScheme } from '@mantine/core';
import { IconEye, IconMapPin, IconPencil } from '@tabler/icons-react';
import * as monaco from 'monaco-editor/editor/editor.api';
import { useEffect, useId, useState } from 'react';

import '../monaco';

type Props = {
  label: string;
  value: string;
  readOnly?: boolean;
  onChange?: (value: string) => void;
  onCursorChange?: (position: CursorPosition) => void;
  remoteCursor?: CursorPosition & { label: string };
};

export type CursorPosition = {
  lineNumber: number;
  column: number;
};

export function CodePane({
  label,
  value,
  readOnly = false,
  onChange,
  onCursorChange,
  remoteCursor,
}: Props) {
  const colorScheme = useComputedColorScheme('dark');
  const widgetID = useId();
  const [editorInstance, setEditorInstance] =
    useState<monaco.editor.IStandaloneCodeEditor | null>(null);
  const remoteLine = remoteCursor?.lineNumber;
  const remoteColumn = remoteCursor?.column;
  const remoteLabel = remoteCursor?.label;

  useEffect(() => {
    if (!editorInstance || !onCursorChange) return undefined;
    const listener = editorInstance.onDidChangeCursorPosition(({ position }) => {
      onCursorChange({ lineNumber: position.lineNumber, column: position.column });
    });
    return () => listener.dispose();
  }, [editorInstance, onCursorChange]);

  useEffect(() => {
    const model = editorInstance?.getModel();
    if (!editorInstance || !model || !remoteLine || !remoteColumn || !remoteLabel) {
      return undefined;
    }

    const lineNumber = Math.min(Math.max(remoteLine, 1), model.getLineCount());
    const column = Math.min(
      Math.max(remoteColumn, 1),
      model.getLineMaxColumn(lineNumber),
    );
    const cursorNode = document.createElement('div');
    cursorNode.className = 'opponent-cursor-widget';
    cursorNode.setAttribute(
      'aria-label',
      `${remoteLabel}: строка ${lineNumber}, столбец ${column}`,
    );
    const cursorLabel = document.createElement('span');
    cursorLabel.className = 'opponent-cursor-widget__label';
    cursorLabel.textContent = remoteLabel;
    cursorNode.append(cursorLabel);

    const widget: monaco.editor.IContentWidget = {
      getId: () => `opponent-cursor-${widgetID}`,
      getDomNode: () => cursorNode,
      getPosition: () => ({
        position: { lineNumber, column },
        preference: [monaco.editor.ContentWidgetPositionPreference.EXACT],
      }),
    };
    editorInstance.addContentWidget(widget);
    return () => editorInstance.removeContentWidget(widget);
  }, [editorInstance, remoteLine, remoteColumn, remoteLabel, value, widgetID]);

  return (
    <Paper withBorder radius="md" style={{ overflow: 'hidden' }}>
      <Group justify="space-between" px="md" py="sm">
        <Text fw={600}>{label}</Text>
        <Group gap="xs">
          {remoteCursor && (
            <Badge color="indigo" variant="light" leftSection={<IconMapPin size={13} />}>
              Курсор {remoteCursor.lineNumber}:{remoteCursor.column}
            </Badge>
          )}
          <Badge
            variant="light"
            color={readOnly ? 'gray' : 'indigo'}
            leftSection={readOnly ? <IconEye size={13} /> : <IconPencil size={13} />}
          >
            {readOnly ? 'Только чтение' : 'Редактирование'}
          </Badge>
        </Group>
      </Group>
      <Editor
        height="360px"
        language="go"
        theme={colorScheme === 'dark' ? 'vs-dark' : 'vs'}
        value={value}
        onMount={setEditorInstance}
        onChange={(nextValue) => onChange?.(nextValue ?? '')}
        options={{
          readOnly,
          automaticLayout: true,
          minimap: { enabled: false },
          fontFamily: 'JetBrains Mono, Consolas, monospace',
          fontSize: 14,
          lineNumbersMinChars: 3,
          padding: { top: 14, bottom: 14 },
          scrollBeyondLastLine: false,
          tabSize: 4,
        }}
      />
    </Paper>
  );
}
