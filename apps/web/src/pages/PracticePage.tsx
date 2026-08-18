import {
  Alert,
  Badge,
  Button,
  Card,
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
import { notifications } from '@mantine/notifications';
import { useMutation, useQuery } from '@tanstack/react-query';
import { IconAlertCircle, IconPlayerPlay, IconSearch } from '@tabler/icons-react';
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import {
  ApiError,
  listPracticeProblems,
  startPracticeSession,
  type PracticeProblem,
  type ProblemClass,
  type User,
} from '../api/client';
import { difficultyColor, difficultyLabel } from '../difficulties';
import { problemClassColor, problemClassLabel, problemClassOptions } from '../problemClasses';

export function PracticePage({ currentUser }: { currentUser: User }) {
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [problemClass, setProblemClass] = useState<ProblemClass | 'all'>('all');
  const [difficulty, setDifficulty] = useState<'all' | PracticeProblem['difficulty']>('all');
  const [solvedFilter, setSolvedFilter] = useState<'all' | 'unsolved' | 'solved'>('all');

  const problemsQuery = useQuery({
    queryKey: ['practice-problems', currentUser.id],
    queryFn: listPracticeProblems,
  });
  const startMutation = useMutation({
    mutationFn: startPracticeSession,
    onSuccess: (session) => navigate(`/practice/${session.id}`),
    onError: (error) => {
      notifications.show({
        title: 'Не удалось открыть задачу',
        message: error instanceof ApiError ? error.message : 'Проверьте соединение с сервером',
        color: 'red',
      });
    },
  });

  const filtered = useMemo(() => {
    const items = problemsQuery.data ?? [];
    const needle = query.trim().toLowerCase();
    return items.filter((problem) => {
      if (problemClass !== 'all' && problem.problem_class !== problemClass) return false;
      if (difficulty !== 'all' && problem.difficulty !== difficulty) return false;
      if (solvedFilter === 'solved' && !problem.solved) return false;
      if (solvedFilter === 'unsolved' && problem.solved) return false;
      if (!needle) return true;
      return (
        problem.title.toLowerCase().includes(needle) || problem.slug.toLowerCase().includes(needle)
      );
    });
  }, [difficulty, problemClass, problemsQuery.data, query, solvedFilter]);

  const solvedCount = (problemsQuery.data ?? []).filter((problem) => problem.solved).length;
  const totalCount = problemsQuery.data?.length ?? 0;

  return (
    <Container size="lg">
      <Stack gap="lg">
        <Paper withBorder p={{ base: 'lg', sm: 'xl' }} radius="lg">
          <Badge variant="light" color="teal" mb="sm">
            Одиночный режим
          </Badge>
          <Title order={1}>Тренировка</Title>
          <Text c="dimmed" mt="sm" maw={680}>
            Решайте задачи без соперника. Черновик и отметка «решено» сохраняются на сервере.
          </Text>
          {problemsQuery.isSuccess && (
            <Text mt="md" fw={600}>
              Решено {solvedCount} из {totalCount}
            </Text>
          )}
        </Paper>

        <Stack gap="sm">
          <SegmentedControl
            value={problemClass}
            onChange={(value) => setProblemClass(value as ProblemClass | 'all')}
            data={[{ label: 'Все топики', value: 'all' }, ...problemClassOptions]}
            aria-label="Фильтр по классу задач"
            fullWidth
          />
          <Group wrap="wrap">
            <SegmentedControl
              value={difficulty}
              onChange={(value) => setDifficulty(value as typeof difficulty)}
              data={[
                { label: 'Любая сложность', value: 'all' },
                { label: difficultyLabel.easy, value: 'easy' },
                { label: difficultyLabel.medium, value: 'medium' },
                { label: difficultyLabel.hard, value: 'hard' },
              ]}
              aria-label="Фильтр по сложности"
            />
            <SegmentedControl
              value={solvedFilter}
              onChange={(value) => setSolvedFilter(value as typeof solvedFilter)}
              data={[
                { label: 'Все', value: 'all' },
                { label: 'Нерешённые', value: 'unsolved' },
                { label: 'Решённые', value: 'solved' },
              ]}
              aria-label="Фильтр по статусу решения"
            />
            <TextInput
              value={query}
              onChange={(event) => setQuery(event.currentTarget.value)}
              leftSection={<IconSearch size={16} />}
              placeholder="Поиск по названию"
              aria-label="Поиск задач"
              w={{ base: '100%', sm: 280 }}
              ml="auto"
            />
          </Group>
        </Stack>

        {problemsQuery.isLoading && <Loader mx="auto" aria-label="Загрузка задач" />}

        {problemsQuery.isError && (
          <Alert icon={<IconAlertCircle size={18} />} color="red">
            Не удалось загрузить каталог задач.
          </Alert>
        )}

        <SimpleGrid cols={{ base: 1, sm: 2 }}>
          {filtered.map((problem) => (
            <Card key={problem.slug} withBorder padding="lg" radius="md">
              <Stack gap="sm">
                <Group gap="xs">
                  <Badge color={problemClassColor[problem.problem_class]} variant="light">
                    {problemClassLabel[problem.problem_class]}
                  </Badge>
                  <Badge color={difficultyColor[problem.difficulty]} variant="light">
                    {difficultyLabel[problem.difficulty]}
                  </Badge>
                  {problem.solved && (
                    <Badge color="green" variant="filled">
                      Решено
                    </Badge>
                  )}
                </Group>
                <Title order={3}>{problem.title}</Title>
                <Button
                  mt="auto"
                  leftSection={<IconPlayerPlay size={16} />}
                  loading={startMutation.isPending && startMutation.variables === problem.slug}
                  onClick={() => startMutation.mutate(problem.slug)}
                >
                  Решать
                </Button>
              </Stack>
            </Card>
          ))}
        </SimpleGrid>

        {problemsQuery.isSuccess && filtered.length === 0 && (
          <Paper withBorder p="xl" ta="center">
            <Text c="dimmed">Задачи не найдены</Text>
          </Paper>
        )}
      </Stack>
    </Container>
  );
}
