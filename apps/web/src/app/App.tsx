import {
  Alert,
  AppShell,
  Badge,
  Button,
  Center,
  Container,
  Group,
  Loader,
  Text,
  Title,
} from '@mantine/core';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { IconLogout, IconSwords } from '@tabler/icons-react';
import { lazy, Suspense } from 'react';
import { Link, Navigate, Route, Routes } from 'react-router-dom';

import { getMe, logout, type User } from '../api/client';
import { ThemeToggle } from '../components/ThemeToggle';
import { AuthPage } from '../pages/AuthPage';
import { LobbyPage } from '../pages/LobbyPage';

const MatchShellPage = lazy(() =>
  import('../pages/MatchShellPage').then((module) => ({ default: module.MatchShellPage })),
);

export function App() {
  const queryClient = useQueryClient();
  const meQuery = useQuery({ queryKey: ['me'], queryFn: getMe, retry: false });
  const logoutMutation = useMutation({
    mutationFn: logout,
    onSuccess: () => queryClient.setQueryData<User | null>(['me'], null),
  });

  const currentUser = meQuery.data ?? null;

  return (
    <AppShell header={{ height: 68 }} padding="md">
      <AppShell.Header>
        <Container size="xl" h="100%">
          <Group h="100%" justify="space-between" wrap="nowrap">
            <Link to="/" style={{ color: 'inherit', textDecoration: 'none' }}>
              <Group gap="sm">
                <IconSwords size={28} stroke={1.8} />
                <div>
                  <Group gap="xs">
                    <Title order={3}>CodeBattle</Title>
                    <Badge variant="light" color="indigo">
                      MVP
                    </Badge>
                  </Group>
                  <Text size="xs" c="dimmed" visibleFrom="sm">
                    Go-дуэли в реальном времени
                  </Text>
                </div>
              </Group>
            </Link>
            <Group gap="xs" wrap="nowrap">
              {currentUser && (
                <>
                  <Text size="sm" fw={600} visibleFrom="sm">
                    {currentUser.username}
                  </Text>
                  <Button
                    variant="subtle"
                    color="gray"
                    size="compact-sm"
                    leftSection={<IconLogout size={16} />}
                    loading={logoutMutation.isPending}
                    onClick={() => logoutMutation.mutate()}
                  >
                    Выйти
                  </Button>
                </>
              )}
              <ThemeToggle />
            </Group>
          </Group>
        </Container>
      </AppShell.Header>

      <AppShell.Main>
        {meQuery.isLoading && (
          <Center mih="60vh">
            <Loader aria-label="Проверка сессии" />
          </Center>
        )}

        {meQuery.isError && (
          <Container size="sm" py="xl">
            <Alert color="red">API недоступен. Запустите backend и PostgreSQL.</Alert>
          </Container>
        )}

        {!meQuery.isLoading && !meQuery.isError && !currentUser && (
          <AuthPage onAuthenticated={(user) => queryClient.setQueryData(['me'], user)} />
        )}

        {currentUser && (
          <Routes>
            <Route path="/" element={<LobbyPage currentUser={currentUser} />} />
            <Route
              path="/matches/:matchId"
              element={
                <Suspense
                  fallback={
                    <Center mih="60vh">
                      <Loader aria-label="Загрузка редактора" />
                    </Center>
                  }
                >
                  <MatchShellPage currentUser={currentUser} />
                </Suspense>
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        )}
      </AppShell.Main>
    </AppShell>
  );
}
