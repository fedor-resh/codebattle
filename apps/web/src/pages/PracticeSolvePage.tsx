import {
  Alert,
  Badge,
  Button,
  Center,
  Container,
  Group,
  Loader,
  Paper,
  Stack,
  Title,
} from '@mantine/core';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { IconArrowLeft, IconCheck, IconPlayerPlay } from '@tabler/icons-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import {
  ApiError,
  createPracticeSubmission,
  getPracticeSession,
  getSubmission,
  updatePracticeCode,
  type Submission,
  type User,
} from '../api/client';
import { CodePane } from '../components/CodePane';
import { JudgeResultPanel } from '../components/JudgeResultPanel';
import { ProblemPanel } from '../components/ProblemPanel';
import { problemClassColor, problemClassLabel } from '../problemClasses';

export function PracticeSolvePage({ currentUser }: { currentUser: User }) {
  const { sessionId = '' } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [source, setSource] = useState('');
  const [submissionID, setSubmissionID] = useState('');
  const revision = useRef(0);
  const sourceRef = useRef('');
  const loadedSessionID = useRef('');
  const syncTimer = useRef<number | undefined>(undefined);

  const sessionQuery = useQuery({
    queryKey: ['practice-session', sessionId, currentUser.id],
    queryFn: () => getPracticeSession(sessionId),
    enabled: Boolean(sessionId),
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
    mutationFn: () => createPracticeSubmission(sessionId, source),
    onSuccess: (submission) => {
      setSubmissionID(submission.id);
      queryClient.setQueryData(['submission', submission.id], submission);
    },
  });

  useEffect(() => {
    const session = sessionQuery.data;
    if (!session || loadedSessionID.current === session.id) return;
    loadedSessionID.current = session.id;
    revision.current = session.revision;
    const restored = session.source_code || session.problem.starter_code;
    sourceRef.current = restored;
    setSource(restored);
    setSubmissionID('');
  }, [sessionQuery.data]);

  useEffect(() => {
    if (submissionQuery.data?.status === 'accepted') {
      void queryClient.invalidateQueries({ queryKey: ['practice-session', sessionId, currentUser.id] });
      void queryClient.invalidateQueries({ queryKey: ['practice-problems', currentUser.id] });
    }
  }, [currentUser.id, queryClient, sessionId, submissionQuery.data?.status]);

  useEffect(
    () => () => {
      if (syncTimer.current !== undefined) window.clearTimeout(syncTimer.current);
    },
    [],
  );

  const scheduleEditorSync = useCallback(
    (nextSource: string) => {
      revision.current += 1;
      if (syncTimer.current !== undefined) window.clearTimeout(syncTimer.current);
      const currentRevision = revision.current;
      syncTimer.current = window.setTimeout(() => {
        void updatePracticeCode(sessionId, nextSource, currentRevision).catch(() => undefined);
      }, 150);
    },
    [sessionId],
  );

  const handleSourceChange = useCallback(
    (value: string) => {
      sourceRef.current = value;
      setSource(value);
      scheduleEditorSync(value);
    },
    [scheduleEditorSync],
  );

  if (sessionQuery.isLoading) {
    return (
      <Center mih="60vh">
        <Loader aria-label="Загрузка задачи" />
      </Center>
    );
  }

  if (!sessionQuery.data || sessionQuery.isError) {
    return (
      <Container size="sm" py="xl">
        <Alert color="red" title="Сессия недоступна">
          Тренировка не найдена или у вас нет доступа к ней.
        </Alert>
        <Button mt="md" leftSection={<IconArrowLeft size={17} />} onClick={() => navigate('/practice')}>
          К списку задач
        </Button>
      </Container>
    );
  }

  const session = sessionQuery.data;
  const solved = Boolean(session.solved_at) || submissionQuery.data?.status === 'accepted';

  return (
    <Container size="xl" fluid>
      <Stack gap="md">
        <Paper withBorder p="md" radius="md">
          <Group justify="space-between" wrap="wrap">
            <Group>
              <Button
                variant="subtle"
                leftSection={<IconArrowLeft size={17} />}
                onClick={() => navigate('/practice')}
              >
                К списку
              </Button>
              <Title order={3}>{session.problem.title}</Title>
              <Badge color={problemClassColor[session.problem.problem_class]} variant="light">
                {problemClassLabel[session.problem.problem_class]}
              </Badge>
              {solved && (
                <Badge color="green" leftSection={<IconCheck size={13} />}>
                  Решено
                </Badge>
              )}
            </Group>
          </Group>
        </Paper>

        <div className="match-grid">
          <ProblemPanel problem={session.problem} />
          <div className="editor-stack">
            <CodePane
              label="Ваш код"
              value={source}
              onChange={handleSourceChange}
              functionSignature={session.problem.function_signature}
            />
            <Paper withBorder p="md" radius="md">
              <Stack>
                <Button
                  loading={submitMutation.isPending}
                  leftSection={<IconPlayerPlay size={18} />}
                  onClick={() => submitMutation.mutate()}
                  ml="auto"
                >
                  Отправить решение
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
      </Stack>
    </Container>
  );
}
