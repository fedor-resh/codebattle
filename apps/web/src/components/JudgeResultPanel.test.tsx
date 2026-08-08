import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';

import type { Submission } from '../api/client';
import { JudgeResultPanel } from './JudgeResultPanel';

const submission: Submission = {
  id: 'submission-1',
  match_id: 'match-1',
  user_id: 'user-1',
  status: 'wrong_answer',
  created_at: '2026-08-08T12:00:00Z',
  result: {
    message: 'Публичный пример 1 не пройден',
    duration_ms: 124,
    passed_tests: 1,
    total_tests: 2,
    console_output: 'checking 2 + 2\n',
    test_cases: [
      {
        kind: 'public',
        index: 1,
        status: 'failed',
        input: '2 2',
        expected: '4',
        actual_available: true,
      },
      { kind: 'hidden', status: 'passed' },
    ],
  },
};

describe('JudgeResultPanel', () => {
  it('shows a test summary, public values and a protected hidden result', () => {
    render(
      <MantineProvider>
        <JudgeResultPanel submission={submission} />
      </MantineProvider>,
    );

    expect(screen.getByText('Неверный ответ')).toBeInTheDocument();
    expect(screen.getByText('1 из 2')).toBeInTheDocument();
    expect(screen.getByLabelText('Пройдено 1 из 2 проверок')).toBeInTheDocument();
    expect(screen.getByText('Пример 1')).toBeInTheDocument();
    expect(screen.getByText('2 2')).toBeInTheDocument();
    expect(screen.getByText('4')).toBeInTheDocument();
    expect(screen.getByText('""')).toBeInTheDocument();
    expect(screen.getByText('Вывод консоли')).toBeInTheDocument();
    expect(screen.getByText('checking 2 + 2')).toBeInTheDocument();
    expect(screen.getByText('Скрытые тесты')).toBeInTheDocument();
    expect(screen.getByText(/Входные данные скрыты/)).toBeInTheDocument();
  });
});
