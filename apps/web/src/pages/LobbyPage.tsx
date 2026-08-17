import {
  Alert,
  Badge,
  Button,
  Container,
  Group,
  Loader,
  Paper,
  SegmentedControl,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useDebouncedValue } from '@mantine/hooks';
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { IconAlertCircle, IconSearch, IconSwords, IconTarget } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import {
  acceptInvitation,
  ApiError,
  createInvitation,
  declineInvitation,
  getInvitationState,
  heartbeat,
  listUsers,
  type InvitationState,
  type ProblemClass,
  type User,
} from '../api/client';
import { InvitationCountdown } from '../components/InvitationCountdown';
import { InvitationModal } from '../components/InvitationModal';
import { UserListItem } from '../components/UserListItem';
import { problemClassColor, problemClassLabel, problemClassOptions } from '../problemClasses';

function showRequestError(error: unknown) {
  notifications.show({
    title: 'Не удалось выполнить действие',
    message: error instanceof ApiError ? error.message : 'Проверьте соединение с сервером',
    color: 'red',
  });
}

export function LobbyPage({ currentUser }: { currentUser: User }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [query, setQuery] = useState('');
  const [debouncedQuery] = useDebouncedValue(query.trim(), 250);
  const [cursor, setCursor] = useState('');
  const [problemClass, setProblemClass] = useState<ProblemClass>('algorithms');

  const usersQuery = useQuery({
    queryKey: ['users', currentUser.id, debouncedQuery, cursor],
    queryFn: () => listUsers(debouncedQuery, cursor),
    refetchInterval: 20_000,
  });
  const invitationQuery = useQuery({
    queryKey: ['invitation-state', currentUser.id],
    queryFn: getInvitationState,
    refetchInterval: 2_000,
  });

  useEffect(() => setCursor(''), [debouncedQuery]);

  useEffect(() => {
    const matchID = invitationQuery.data?.match?.id;
    if (matchID) navigate(`/matches/${matchID}`);
  }, [invitationQuery.data?.match?.id, navigate]);

  useEffect(() => {
    const interval = window.setInterval(() => {
      void heartbeat()
        .then(() => usersQuery.refetch())
        .catch(() => undefined);
    }, 20_000);
    return () => window.clearInterval(interval);
  }, [usersQuery.refetch]);

  const createMutation = useMutation({
    mutationFn: (receiverID: string) => createInvitation(receiverID, problemClass),
    onSuccess: (invitation) => {
      queryClient.setQueryData<InvitationState>(['invitation-state', currentUser.id], {
        outgoing: invitation,
      });
      notifications.show({
        title: 'Приглашение отправлено',
        message: `${invitation.receiver.username} может принять его в течение 30 секунд`,
        color: 'indigo',
        icon: <IconSwords size={18} />,
      });
    },
    onError: showRequestError,
  });
  const acceptMutation = useMutation({
    mutationFn: (invitationID: string) => acceptInvitation(invitationID),
    onSuccess: (match) => navigate(`/matches/${match.id}`),
    onError: (error) => {
      showRequestError(error);
      void invitationQuery.refetch();
    },
  });
  const declineMutation = useMutation({
    mutationFn: (invitationID: string) => declineInvitation(invitationID),
    onSuccess: () => {
      queryClient.setQueryData<InvitationState>(['invitation-state', currentUser.id], {});
      void usersQuery.refetch();
    },
    onError: showRequestError,
  });

  const invitationState = invitationQuery.data;
  const hasPendingInvitation = Boolean(invitationState?.incoming || invitationState?.outgoing);

  return (
    <Container size="lg">
      <Stack gap="lg">
        <Paper withBorder p={{ base: 'lg', sm: 'xl' }} radius="lg">
          <Badge variant="light" color="green" mb="sm">
            Вы доступны для приглашений
          </Badge>
          <Title order={1}>Найдите соперника и начните Go-дуэль</Title>
          <Text c="dimmed" mt="sm" maw={680}>
            Приглашение действует 30 секунд. После принятия у обоих игроков откроется одна
            комната.
          </Text>
          <Stack gap={6} mt="lg" maw={520}>
            <Text size="sm" fw={700}>Класс задач</Text>
            <SegmentedControl
              value={problemClass}
              onChange={(value) => setProblemClass(value as ProblemClass)}
              data={problemClassOptions}
              disabled={hasPendingInvitation || createMutation.isPending}
              aria-label="Класс задач для дуэли"
              fullWidth
            />
            <Text size="xs" c="dimmed">
              Выбранный класс действует во всех раундах серии.
            </Text>
            <Button
              mt="sm"
              variant="light"
              leftSection={<IconTarget size={18} />}
              onClick={() => navigate('/practice')}
            >
              Решать задачи одному
            </Button>
          </Stack>
        </Paper>

        {invitationState?.outgoing && (
          <Alert color="indigo" icon={<IconSwords size={18} />} title="Приглашение отправлено">
            <Stack gap="sm">
              <Text size="sm">
                Ожидаем ответ от <b>{invitationState.outgoing.receiver.username}</b>
              </Text>
              <Badge
                variant="light"
                color={problemClassColor[invitationState.outgoing.problem_class]}
              >
                {problemClassLabel[invitationState.outgoing.problem_class]}
              </Badge>
              <InvitationCountdown expiresAt={invitationState.outgoing.expires_at} />
            </Stack>
          </Alert>
        )}

        <Group justify="space-between" align="end">
          <div>
            <Title order={2}>Все пользователи</Title>
            <Text size="sm" c="dimmed">
              Доступные игроки отображаются первыми
            </Text>
          </div>
          <TextInput
            value={query}
            onChange={(event) => setQuery(event.currentTarget.value)}
            leftSection={<IconSearch size={16} />}
            placeholder="Поиск по username"
            aria-label="Поиск пользователей"
            w={{ base: '100%', sm: 300 }}
          />
        </Group>

        {usersQuery.isLoading && <Loader mx="auto" aria-label="Загрузка пользователей" />}

        {usersQuery.isError && (
          <Alert icon={<IconAlertCircle size={18} />} color="red">
            Не удалось загрузить пользователей. Проверьте соединение с API.
          </Alert>
        )}

        <SimpleGrid cols={{ base: 1, md: 2 }}>
          {usersQuery.data?.items.map((user) => (
            <UserListItem
              key={user.id}
              user={user}
              isSelf={user.id === currentUser.id}
              inviteDisabled={hasPendingInvitation || createMutation.isPending}
              onInvite={(selectedUser) => createMutation.mutate(selectedUser.id)}
            />
          ))}
        </SimpleGrid>

        {usersQuery.data?.items.length === 0 && (
          <Paper withBorder p="xl" ta="center">
            <Text c="dimmed">Пользователи не найдены</Text>
          </Paper>
        )}

        {usersQuery.data?.next_cursor && (
          <Button
            variant="light"
            mx="auto"
            onClick={() => setCursor(usersQuery.data.next_cursor ?? '')}
          >
            Следующая страница
          </Button>
        )}
      </Stack>

      <InvitationModal
        invitation={invitationState?.incoming}
        loading={acceptMutation.isPending || declineMutation.isPending}
        onAccept={() => {
          if (invitationState?.incoming) acceptMutation.mutate(invitationState.incoming.id);
        }}
        onDecline={() => {
          if (invitationState?.incoming) declineMutation.mutate(invitationState.incoming.id);
        }}
      />
    </Container>
  );
}
