import Editor from '@monaco-editor/react';
import { Badge, Group, Paper, Text, useComputedColorScheme } from '@mantine/core';
import { IconEye, IconPencil } from '@tabler/icons-react';

import '../monaco';

type Props = {
  label: string;
  value: string;
  readOnly?: boolean;
  onChange?: (value: string) => void;
};

export function CodePane({ label, value, readOnly = false, onChange }: Props) {
  const colorScheme = useComputedColorScheme('dark');

  return (
    <Paper withBorder radius="md" style={{ overflow: 'hidden' }}>
      <Group justify="space-between" px="md" py="sm">
        <Text fw={600}>{label}</Text>
        <Badge
          variant="light"
          color={readOnly ? 'gray' : 'indigo'}
          leftSection={readOnly ? <IconEye size={13} /> : <IconPencil size={13} />}
        >
          {readOnly ? 'Только чтение' : 'Редактирование'}
        </Badge>
      </Group>
      <Editor
        height="360px"
        language="go"
        theme={colorScheme === 'dark' ? 'vs-dark' : 'vs'}
        value={value}
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
