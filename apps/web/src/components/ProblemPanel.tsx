import { Accordion, Badge, Code, Group, Paper, ScrollArea, Stack, Text, Title } from '@mantine/core';

import type { Problem } from '../api/client';
import { difficultyColor, difficultyLabel } from '../difficulties';
const requirementLabel = {
  goroutine: 'goroutine',
  channel: 'channel',
  wait_group: 'sync.WaitGroup',
  mutex: 'sync.Mutex',
  select: 'select',
  context_cancel: 'context.WithCancel',
} as const;

function formatValue(value: unknown): string {
  return JSON.stringify(value, null, 2) ?? String(value);
}

export function ProblemPanel({ problem }: { problem: Problem }) {
  const activeRequirements = (Object.keys(requirementLabel) as Array<keyof typeof requirementLabel>)
    .filter((requirement) => problem.requirements[requirement]);

  return (
    <Paper withBorder p="lg" h="100%">
      <ScrollArea h="100%" offsetScrollbars>
        <Stack gap="md" pr="sm">
          <div>
            <Group gap="xs" mb={4}>
              <Badge color={difficultyColor[problem.difficulty]} variant="light">
                {difficultyLabel[problem.difficulty]}
              </Badge>
              <Text size="xs" c="dimmed">
                {problem.time_limit_ms} мс · {problem.memory_limit_mb} МБ
              </Text>
            </Group>
            <Title order={2}>{problem.title}</Title>
          </div>
          <Text style={{ whiteSpace: 'pre-wrap' }}>{problem.statement_markdown.replace(/^# .+\n+/, '')}</Text>
          <Code block>{problem.function_signature}</Code>
          {activeRequirements.length > 0 && (
            <div>
              <Text size="sm" fw={700} mb={6}>Рекомендуемые инструменты</Text>
              <Group gap="xs">
                {activeRequirements.map((requirement) => (
                  <Badge key={requirement} color="violet" variant="light">
                    {requirementLabel[requirement]}
                  </Badge>
                ))}
              </Group>
            </div>
          )}
          <Accordion variant="contained" defaultValue="example-0">
            {problem.public_tests.map((example, index) => (
              <Accordion.Item key={`example-${index}`} value={`example-${index}`}>
                <Accordion.Control>Пример {index + 1}</Accordion.Control>
                <Accordion.Panel>
                  <Code block>{`Аргументы:\n${formatValue(example.arguments ?? [example.input])}\n\nРезультат:\n${formatValue(example.expected)}`}</Code>
                </Accordion.Panel>
              </Accordion.Item>
            ))}
          </Accordion>
        </Stack>
      </ScrollArea>
    </Paper>
  );
}
