import { Alert, Code, Group, Loader, Paper, Stack, Text } from '@mantine/core';
import { IconAlertTriangle, IconCheck, IconClock, IconX } from '@tabler/icons-react';

import type { Submission } from '../api/client';

const pendingStatuses = new Set(['queued', 'compiling', 'running']);

const labels: Record<Submission['status'], string> = {
  queued: 'В очереди',
  compiling: 'Компиляция',
  running: 'Выполнение тестов',
  accepted: 'Решение принято',
  wrong_answer: 'Неверный ответ',
  compile_error: 'Ошибка компиляции',
  runtime_error: 'Ошибка выполнения',
  time_limit: 'Превышен лимит времени',
  memory_limit: 'Превышен лимит памяти',
  internal_error: 'Ошибка judge',
};

export function JudgeResultPanel({ submission }: { submission?: Submission }) {
  if (!submission) {
    return (
      <Paper withBorder p="md" radius="md">
        <Text fw={600}>Проверка решения</Text>
        <Text size="sm" c="dimmed">
          Отправьте код, чтобы запустить public и hidden-тесты
        </Text>
      </Paper>
    );
  }

  if (pendingStatuses.has(submission.status)) {
    return (
      <Alert color="yellow" icon={<Loader size={18} />} title={labels[submission.status]}>
        Сервер безопасно собирает и проверяет решение.
      </Alert>
    );
  }

  const accepted = submission.status === 'accepted';
  const internal = submission.status === 'internal_error';
  const Icon = accepted ? IconCheck : internal ? IconAlertTriangle : IconX;
  return (
    <Alert color={accepted ? 'green' : internal ? 'yellow' : 'red'} icon={<Icon size={18} />}>
      <Stack gap="xs">
        <Group justify="space-between">
          <Text fw={700}>{labels[submission.status]}</Text>
          {submission.result?.duration_ms !== undefined && (
            <Group gap={4}>
              <IconClock size={14} />
              <Text size="xs">{submission.result.duration_ms} мс</Text>
            </Group>
          )}
        </Group>
        {submission.result?.message && <Code block>{submission.result.message}</Code>}
      </Stack>
    </Alert>
  );
}
