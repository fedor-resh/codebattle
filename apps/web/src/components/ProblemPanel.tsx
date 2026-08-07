import { Accordion, Badge, Code, Group, Paper, ScrollArea, Stack, Text, Title } from '@mantine/core';

import type { Problem } from '../api/client';

const difficultyLabel = { easy: 'Простая', medium: 'Средняя', hard: 'Сложная' };
const difficultyColor = { easy: 'green', medium: 'yellow', hard: 'red' };

export function ProblemPanel({ problem }: { problem: Problem }) {
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
          <Accordion variant="contained" defaultValue="example-0">
            {problem.public_tests.map((example, index) => (
              <Accordion.Item key={`${example.input}-${index}`} value={`example-${index}`}>
                <Accordion.Control>Пример {index + 1}</Accordion.Control>
                <Accordion.Panel>
                  <Code block>{`Input:  ${JSON.stringify(example.input)}\nOutput: ${JSON.stringify(example.expected)}`}</Code>
                </Accordion.Panel>
              </Accordion.Item>
            ))}
          </Accordion>
        </Stack>
      </ScrollArea>
    </Paper>
  );
}
