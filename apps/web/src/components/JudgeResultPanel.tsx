import {
  Alert,
  Badge,
  Code,
  Group,
  Loader,
  Paper,
  Progress,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
} from '@mantine/core';
import {
  IconAlertTriangle,
  IconCheck,
  IconClock,
  IconEyeOff,
  IconFlask,
  IconPlayerPause,
  IconTerminal2,
  IconX,
} from '@tabler/icons-react';

import type { Submission, SubmissionTestCase } from '../api/client';

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

const testStatuses = {
  passed: { label: 'Пройден', color: 'green', Icon: IconCheck },
  failed: { label: 'Ошибка', color: 'red', Icon: IconX },
  not_run: { label: 'Не завершён', color: 'yellow', Icon: IconPlayerPause },
} as const;

function TestValue({ label, value }: { label: string; value?: string }) {
  return (
    <Stack gap={4} miw={0}>
      <Text size="xs" fw={600} c="dimmed">
        {label}
      </Text>
      <Code block style={{ minHeight: 42, whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
        {value === undefined ? 'Функция не вернула значение' : value === '' ? '""' : value}
      </Code>
    </Stack>
  );
}

function TestCaseCard({ testCase }: { testCase: SubmissionTestCase }) {
  const status = testStatuses[testCase.status];
  const StatusIcon = status.Icon;
  const hidden = testCase.kind === 'hidden';

  return (
    <Paper withBorder p="sm" radius="md">
      <Stack gap="sm">
        <Group justify="space-between" align="flex-start" wrap="nowrap">
          <Group gap="sm" wrap="nowrap">
            <ThemeIcon color={status.color} variant="light" radius="xl" size="lg">
              <StatusIcon size={18} aria-hidden="true" />
            </ThemeIcon>
            <div>
              <Text fw={650}>{hidden ? 'Скрытые тесты' : `Пример ${testCase.index}`}</Text>
              <Text size="xs" c="dimmed">
                {hidden ? 'Защищённая группа проверок' : 'Публичные тестовые данные'}
              </Text>
            </div>
          </Group>
          <Badge color={status.color} variant="light" leftSection={<StatusIcon size={12} />}>
            {status.label}
          </Badge>
        </Group>

        {hidden ? (
          <Group gap="xs" wrap="nowrap">
            <IconEyeOff size={16} aria-hidden="true" />
            <Text size="sm" c="dimmed">
              Входные данные скрыты. Показывается только итог проверки.
            </Text>
          </Group>
        ) : (
          <>
            <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="sm">
              <TestValue label="Аргументы" value={testCase.input ?? ''} />
              <TestValue label="Ожидалось" value={testCase.expected ?? ''} />
              <TestValue
                label={testCase.actual_truncated ? 'Получено (сокращено)' : 'Получено'}
                value={testCase.actual_available ? (testCase.actual ?? '') : undefined}
              />
            </SimpleGrid>
            {!testCase.actual_available && testCase.status === 'passed' && (
              <Text size="xs" c="dimmed">
                Фактический вывод не попал в отчёт, но проверка завершилась успешно.
              </Text>
            )}
            {testCase.error && (
              <Alert color="red" variant="light" title="Ошибка выполнения">
                <Code block style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
                  {testCase.error}
                </Code>
              </Alert>
            )}
          </>
        )}
      </Stack>
    </Paper>
  );
}

export function JudgeResultPanel({ submission }: { submission?: Submission }) {
  if (!submission) {
    return (
      <Paper withBorder p="md" radius="md">
        <Group gap="sm" align="flex-start">
          <ThemeIcon color="blue" variant="light" radius="xl">
            <IconFlask size={16} aria-hidden="true" />
          </ThemeIcon>
          <div>
            <Text fw={600}>Проверка решения</Text>
            <Text size="sm" c="dimmed">
              Отправьте код, чтобы увидеть результаты публичных и скрытых тестов.
            </Text>
          </div>
        </Group>
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
  const testCases = submission.result?.test_cases ?? [];
  const totalTests = submission.result?.total_tests ?? testCases.length;
  const passedTests =
    submission.result?.passed_tests ?? testCases.filter((testCase) => testCase.status === 'passed').length;
  const failed = testCases.some((testCase) => testCase.status === 'failed');
  const progressColor = failed ? 'red' : passedTests === totalTests ? 'green' : 'yellow';
  const progress = totalTests > 0 ? (passedTests / totalTests) * 100 : 0;
  const technicalMessage = ['compile_error', 'runtime_error', 'internal_error'].includes(
    submission.status,
  );
  const hasConsoleReport = submission.result?.console_output !== undefined;
  const consoleOutput = submission.result?.console_output ?? '';

  return (
    <Alert color={accepted ? 'green' : internal ? 'yellow' : 'red'} icon={<Icon size={18} />}>
      <Stack gap="md">
        <Stack gap={6}>
          <Group justify="space-between" align="center">
            <Text fw={700}>{labels[submission.status]}</Text>
            {submission.result?.duration_ms !== undefined && (
              <Badge
                color="gray"
                variant="light"
                leftSection={<IconClock size={13} aria-hidden="true" />}
              >
                {submission.result.duration_ms} мс
              </Badge>
            )}
          </Group>
          {submission.result?.message &&
            (technicalMessage ? (
              <Code block style={{ whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>
                {submission.result.message}
              </Code>
            ) : (
              <Text size="sm">{submission.result.message}</Text>
            ))}
        </Stack>

        {hasConsoleReport && (
          <Paper withBorder p="sm" radius="md">
            <Stack gap="xs">
              <Group justify="space-between" gap="xs">
                <Group gap="xs">
                  <ThemeIcon color="gray" variant="light" radius="xl" size="md">
                    <IconTerminal2 size={15} aria-hidden="true" />
                  </ThemeIcon>
                  <Text size="sm" fw={650}>
                    Вывод консоли
                  </Text>
                </Group>
                <Badge color="gray" variant="light">
                  stdout / stderr
                </Badge>
              </Group>
              <Code
                block
                style={{
                  maxHeight: 220,
                  overflow: 'auto',
                  whiteSpace: 'pre-wrap',
                  overflowWrap: 'anywhere',
                }}
              >
                {consoleOutput ||
                  (submission.status === 'compile_error'
                    ? 'Код не запускался.'
                    : 'Код ничего не вывел в консоль.')}
              </Code>
              {submission.result?.console_output_truncated && (
                <Text size="xs" c="dimmed">
                  Вывод сокращён из-за ограничения размера.
                </Text>
              )}
            </Stack>
          </Paper>
        )}

        {testCases.length > 0 && (
          <Stack gap="sm">
            <Group justify="space-between">
              <Text size="sm" fw={650}>
                Результаты тестов
              </Text>
              <Text size="sm" fw={650} c={progressColor}>
                {passedTests} из {totalTests}
              </Text>
            </Group>
            <Progress
              value={progress}
              color={progressColor}
              size="md"
              radius="xl"
              aria-label={`Пройдено ${passedTests} из ${totalTests} проверок`}
            />
            <Stack gap="xs">
              {testCases.map((testCase, index) => (
                <TestCaseCard
                  key={`${testCase.kind}-${testCase.index ?? index}`}
                  testCase={testCase}
                />
              ))}
            </Stack>
          </Stack>
        )}
      </Stack>
    </Alert>
  );
}
