import { Button, Center, Stack, Text, Title } from '@mantine/core';
import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <Center mih="70vh">
      <Stack align="center">
        <Title order={1}>404</Title>
        <Text c="dimmed">Такой страницы в CodeBattle нет.</Text>
        <Button component={Link} to="/">
          Вернуться в lobby
        </Button>
      </Stack>
    </Center>
  );
}

