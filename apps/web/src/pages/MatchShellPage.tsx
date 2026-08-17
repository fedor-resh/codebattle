import {
  Alert,
  Avatar,
  Badge,
  Button,
  Center,
  Container,
  Group,
  Loader,
  Paper,
  Stack,
  Tabs,
  Text,
  Title,
} from '@mantine/core';
import { modals } from '@mantine/modals';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { IconArrowLeft, IconCloudCheck, IconDoorExit, IconPlayerPlay } from '@tabler/icons-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import {
  ApiError,
  createSubmission,
  getMatch,
  getSubmission,
  leaveMatch,
  readyForNextRound,
  updateCode,
  type Submission,
  type User,
} from '../api/client';
import { CodePane, type CursorPosition } from '../components/CodePane';
import { JudgeResultPanel } from '../components/JudgeResultPanel';
import { ProblemPanel } from '../components/ProblemPanel';
import { ReadyOverlay } from '../components/ReadyOverlay';
import { problemClassColor, problemClassLabel } from '../problemClasses';

export function MatchShellPage({ currentUser }: { currentUser: User }) {
  const { matchId = '' } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [source, setSource] = useState('');
  const [submissionID, setSubmissionID] = useState('');
  const revision = useRef(0);
  const sourceRef = useRef('');
  const cursorRef = useRef<CursorPosition>({ lineNumber: 1, column: 1 });
  const activeProblemID = useRef('');
  const syncTimer = useRef<number | undefined>(undefined);
  const matchQuery = useQuery({
    queryKey: ['match', matchId],
    queryFn: () => getMatch(matchId),
    enabled: Boolean(matchId),
    refetchInterval: 500,
  });
  const leaveMutation = useMutation({
    mutationFn: () => leaveMatch(matchId),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: ['invitation-state', currentUser.id] });
      void queryClient.invalidateQueries({ queryKey: ['users'] });
      navigate('/');
    },
  });
  const submissionQuery = useQuery({
    queryKey: ['submission', submissionID],
    queryFn: () => getSubmission(submissionID),
    enabled: Boolean(submissionID),
    refetchInterval: (query) => {
      const status = (query.state.data as Submission | undefined)?.status;
      return status && !['queued', 'compiling', 'running'].includes(status) ? false : 1_000;
    },
  });
  const submitMutation = useMutation({
    mutationFn: () => createSubmission(matchId, source),
    onSuccess: (submission) => {
      setSubmissionID(submission.id);
      queryClient.setQueryData(['submission', submission.id], submission);
    },
  });
  const readyMutation = useMutation({
    mutationFn: () => readyForNextRound(matchId),
    onSuccess: (nextMatch) => queryClient.setQueryData(['match', matchId], nextMatch),
  });

  useEffect(() => {
    const problem = matchQuery.data?.problem;
    if (problem && activeProblemID.current !== problem.id) {
      const snapshot = matchQuery.data?.code_snapshots.find(
        (item) => item.user_id === currentUser.id && item.problem_version_id === problem.id,
      );
      activeProblemID.current = problem.id;
      revision.current = snapshot?.revision ?? 0;
      const restoredSource = snapshot?.source_code ?? problem.starter_code;
      sourceRef.current = restoredSource;
      cursorRef.current = {
        lineNumber: snapshot?.cursor_line ?? 1,
        column: snapshot?.cursor_column ?? 1,
      };
      setSource(restoredSource);
      setSubmissionID('');
    }
  }, [currentUser.id, matchQuery.data]);

  useEffect(
    () => () => {
      if (syncTimer.current !== undefined) window.clearTimeout(syncTimer.current);
    },
    [],
  );

  const scheduleEditorSync = useCallback(
    (nextSource: string, cursor: CursorPosition) => {
      revision.current += 1;
      if (syncTimer.current !== undefined) window.clearTimeout(syncTimer.current);
      const currentRevision = revision.current;
      syncTimer.current = window.setTimeout(() => {
        void updateCode(
          matchId,
          nextSource,
          currentRevision,
          cursor.lineNumber,
          cursor.column,
        ).catch(() => undefined);
      }, 150);
    },
    [matchId],
  );

  const handleSourceChange = useCallback(
    (value: string) => {
      sourceRef.current = value;
      setSource(value);
      scheduleEditorSync(value, cursorRef.current);
    },
    [scheduleEditorSync],
  );

  const handleCursorChange = useCallback(
    (position: CursorPosition) => {
      if (
        cursorRef.current.lineNumber === position.lineNumber &&
        cursorRef.current.column === position.column
      ) {
        return;
      }
      cursorRef.current = position;
      scheduleEditorSync(sourceRef.current, position);
    },
    [scheduleEditorSync],
  );

  if (matchQuery.isLoading) {
    return (
      <Center mih="60vh">
        <Loader aria-label="Загрузка комнаты" />
      </Center>
    );
  }

  if (!matchQuery.data || matchQuery.isError) {
    return (
      <Container size="sm" py="xl">
        <Alert color="red" title="Комната недоступна">
          Матч не найден или у вас нет доступа к нему.
        </Alert>
        <Button mt="md" leftSection={<IconArrowLeft size={17} />} onClick={() => navigate('/')}>
          Вернуться в lobby
        </Button>
      </Container>
    );
  }

  const match = matchQuery.data;
  const currentIsPlayerOne = match.player_one.id === currentUser.id;
  const me = currentIsPlayerOne ? match.player_one : match.player_two;
  const opponent = currentIsPlayerOne ? match.player_two : match.player_one;
  const myScore = currentIsPlayerOne ? match.player_one_score : match.player_two_score;
  const opponentScore = currentIsPlayerOne ? match.player_two_score : match.player_one_score;
  const ended = match.state === 'ended';
  const waitingReady = match.state === 'waiting_ready';
  const opponentSnapshot = match.code_snapshots.find(
    (snapshot) =>
      snapshot.user_id === opponent.id && snapshot.problem_version_id === match.problem?.id,
  );
  const currentReady = currentIsPlayerOne ? match.player_one_ready : match.player_two_ready;
  const opponentReady = currentIsPlayerOne ? match.player_two_ready : match.player_one_ready;
  const winner = match.round_winner_id === me.id ? me.username : opponent.username;
  const remoteCursor = opponentSnapshot
    ? {
        lineNumber: opponentSnapshot.cursor_line ?? 1,
        column: opponentSnapshot.cursor_column ?? 1,
        label: opponent.username,
      }
    : undefined;

  const confirmLeave = () =>
    modals.openConfirmModal({
      title: 'Покинуть серию?',
      children: <Text size="sm">Матч завершится для обоих игроков.</Text>,
      labels: { confirm: 'Покинуть', cancel: 'Остаться' },
      confirmProps: { color: 'red' },
      onConfirm: () => leaveMutation.mutate(),
    });

  return (
    <Container size="xl" fluid>
      <Stack gap="md">
        <Paper withBorder p="md" radius="md">
          <Group justify="space-between" wrap="wrap">
            <Group>
              <Title order={3}>Раунд {match.round_number} · {opponent.username}</Title>
              <Badge
                color={problemClassColor[match.problem_class]}
                variant="light"
              >
                {problemClassLabel[match.problem_class]}
              </Badge>
              <Badge
                color={ended ? 'gray' : 'green'}
                variant="light"
                leftSection={<IconCloudCheck size={13} />}
              >
                {ended ? 'Серия завершена' : 'Комната активна'}
              </Badge>
            </Group>
            <Group gap="lg">
              <Group gap="xs">
                <Avatar size="sm" color="indigo">
                  {me.username.slice(0, 2).toUpperCase()}
                </Avatar>
                <Text fw={700}>
                  {me.username} · {myScore}
                </Text>
              </Group>
              <Text c="dimmed">—</Text>
              <Group gap="xs">
                <Text fw={700}>
                  {opponentScore} · {opponent.username}
                </Text>
                <Avatar size="sm" color="green">
                  {opponent.username.slice(0, 2).toUpperCase()}
                </Avatar>
              </Group>
              {ended ? (
                <Button variant="light" onClick={() => navigate('/')}>
                  В lobby
                </Button>
              ) : (
                <Button
                  color="red"
                  variant="subtle"
                  leftSection={<IconDoorExit size={17} />}
                  loading={leaveMutation.isPending}
                  onClick={confirmLeave}
                >
                  Покинуть серию
                </Button>
              )}
            </Group>
          </Group>
        </Paper>

        <Alert color="blue" variant="light">
          Задача зафиксирована для раунда. Код соперника обновляется автоматически.
        </Alert>

        {waitingReady && (
          <ReadyOverlay
            winner={winner}
            winningSource={match.winning_source_code}
            currentReady={currentReady}
            opponentReady={opponentReady}
            loading={readyMutation.isPending}
            onReady={() => readyMutation.mutate()}
          />
        )}

        {match.problem ? (
        <div className="match-grid">
          <ProblemPanel problem={match.problem} />
          <div className="editor-stack">
            <div className="desktop-editors visible-from-lg">
              <CodePane
                label="Ваш код"
                value={source}
                onChange={handleSourceChange}
                onCursorChange={handleCursorChange}
                readOnly={ended}
                functionSignature={match.problem.function_signature}
              />
              <CodePane
                label={`Код ${opponent.username}`}
                value={opponentSnapshot?.source_code ?? match.problem.starter_code}
                remoteCursor={remoteCursor}
                readOnly
              />
            </div>

            <Tabs defaultValue="mine" hiddenFrom="lg">
              <Tabs.List grow>
                <Tabs.Tab value="mine">Ваш код</Tabs.Tab>
                <Tabs.Tab value="opponent">Код соперника</Tabs.Tab>
              </Tabs.List>
              <Tabs.Panel value="mine" pt="md">
                <CodePane
                  label="Ваш код"
                  value={source}
                  onChange={handleSourceChange}
                  onCursorChange={handleCursorChange}
                  readOnly={ended}
                  functionSignature={match.problem.function_signature}
                />
              </Tabs.Panel>
              <Tabs.Panel value="opponent" pt="md">
                <CodePane
                  label={`Код ${opponent.username}`}
                  value={opponentSnapshot?.source_code ?? match.problem.starter_code}
                  remoteCursor={remoteCursor}
                  readOnly
                />
              </Tabs.Panel>
            </Tabs>

            <Paper withBorder p="md" radius="md">
              <Stack>
                <Button
                  disabled={ended}
                  loading={submitMutation.isPending}
                  leftSection={<IconPlayerPlay size={18} />}
                  onClick={() => submitMutation.mutate()}
                  ml="auto"
                >
                  {waitingReady ? 'Проверить вне зачёта' : 'Отправить решение'}
                </Button>
                <JudgeResultPanel submission={submissionQuery.data} />
                {submitMutation.error && (
                  <Alert color="red">
                    {submitMutation.error instanceof ApiError
                      ? submitMutation.error.message
                      : 'Не удалось отправить решение'}
                  </Alert>
                )}
              </Stack>
            </Paper>
          </div>
        </div>
        ) : (
          <Alert color="yellow">Для комнаты не назначена задача. Запустите problem-seed.</Alert>
        )}
      </Stack>
    </Container>
  );
}
