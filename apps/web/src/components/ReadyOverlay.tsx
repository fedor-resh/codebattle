import { Accordion, Alert, Button, Code, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { IconCheck, IconTrophy } from '@tabler/icons-react';

export function ReadyOverlay({
  winner,
  winningSource,
  currentReady,
  opponentReady,
  loading,
  onReady,
}: {
  winner: string;
  winningSource?: string;
  currentReady: boolean;
  opponentReady: boolean;
  loading: boolean;
  onReady: () => void;
}) {
  return (
    <Paper withBorder p="lg" radius="md">
      <Stack>
        <Group>
          <IconTrophy color="var(--mantine-color-yellow-5)" size={32} />
          <div>
            <Title order={3}>Раунд завершён</Title>
            <Text>
              Победитель: <b>{winner}</b>
            </Text>
          </div>
        </Group>
        {winningSource && (
          <Accordion variant="contained">
            <Accordion.Item value="solution">
              <Accordion.Control>Победившее решение</Accordion.Control>
              <Accordion.Panel>
                <Code block>{winningSource}</Code>
              </Accordion.Panel>
            </Accordion.Item>
          </Accordion>
        )}
        <Alert color="indigo" variant="light">
          Можно продолжить редактировать и проверять решения вне зачёта. Результат раунда и счёт
          уже зафиксированы.
        </Alert>
        <Alert color={opponentReady ? 'green' : 'blue'}>
          {opponentReady ? 'Соперник готов к следующей задаче.' : 'Ожидаем готовность соперника.'}
        </Alert>
        <Button
          leftSection={<IconCheck size={18} />}
          disabled={currentReady}
          loading={loading}
          onClick={onReady}
        >
          {currentReady ? 'Вы готовы' : 'Готов к следующей задаче'}
        </Button>
      </Stack>
    </Paper>
  );
}
