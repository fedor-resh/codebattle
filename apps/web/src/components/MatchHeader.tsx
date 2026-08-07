import { Avatar, Badge, Group, Paper, Stack, Text } from '@mantine/core';
import { IconWifi } from '@tabler/icons-react';

export function MatchHeader() {
  return (
    <Paper withBorder p="sm" mb="md">
      <Group justify="space-between" wrap="wrap">
        <Group>
          <Avatar color="indigo">AL</Avatar>
          <Stack gap={0}>
            <Text fw={700}>alice</Text>
            <Text size="xs" c="dimmed">Вы</Text>
          </Stack>
          <Text fw={900} fz="xl">2 : 1</Text>
          <Avatar color="cyan">BO</Avatar>
          <Stack gap={0}>
            <Text fw={700}>bob</Text>
            <Text size="xs" c="dimmed">Соперник</Text>
          </Stack>
        </Group>
        <Group>
          <Badge color="indigo" variant="light">Раунд 4</Badge>
          <Badge color="green" leftSection={<IconWifi size={13} />}>Соединение стабильно</Badge>
        </Group>
      </Group>
    </Paper>
  );
}

