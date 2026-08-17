import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';

import { AppProviders } from './AppProviders';
import { App } from './App';
import { createInvitation } from '../api/client';

vi.mock('../api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/client')>();
  return {
    ...actual,
    getMe: vi.fn().mockResolvedValue({
      id: 'current-user',
      username: 'algo_fox',
      status: 'online_available',
    }),
    listUsers: vi.fn().mockResolvedValue({
      items: [
        { id: 'current-user', username: 'algo_fox', status: 'online_available' },
        { id: 'opponent', username: 'go_ninja', status: 'online_available' },
      ],
    }),
    heartbeat: vi.fn().mockResolvedValue(undefined),
    getInvitationState: vi.fn().mockResolvedValue({}),
    listDuelOptions: vi.fn().mockResolvedValue([
      { problem_class: 'algorithms', difficulty: 'easy', count: 15 },
      { problem_class: 'algorithms', difficulty: 'medium', count: 10 },
      { problem_class: 'concurrency', difficulty: 'easy', count: 2 },
      { problem_class: 'concurrency', difficulty: 'medium', count: 3 },
      { problem_class: 'concurrency', difficulty: 'hard', count: 2 },
      { problem_class: 'oop', difficulty: 'easy', count: 4 },
      { problem_class: 'oop', difficulty: 'medium', count: 9 },
      { problem_class: 'oop', difficulty: 'hard', count: 1 },
    ]),
    createInvitation: vi.fn().mockResolvedValue({
      id: 'invitation-1',
      sender: { id: 'current-user', username: 'algo_fox' },
      receiver: { id: 'opponent', username: 'go_ninja' },
      status: 'pending',
      problem_class: 'concurrency',
      expires_at: '2026-08-13T15:30:00Z',
    }),
  };
});

describe('App', () => {
  it('renders users received from the API and marks the current user', async () => {
    render(
      <AppProviders>
        <MemoryRouter>
          <App />
        </MemoryRouter>
      </AppProviders>,
    );

    expect(
      await screen.findByRole('heading', { name: 'Найдите соперника и начните Go-дуэль' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('go_ninja')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Пригласить algo_fox' })).toBeDisabled();
  });

  it('sends and locks the selected problem class with an invitation', async () => {
    render(
      <AppProviders>
        <MemoryRouter>
          <App />
        </MemoryRouter>
      </AppProviders>,
    );

    await screen.findByText('go_ninja');
    fireEvent.click(screen.getByText('Потоки и горутины'));
    fireEvent.click(screen.getByRole('button', { name: 'Пригласить go_ninja' }));

    await waitFor(() => {
      expect(createInvitation).toHaveBeenCalledWith('opponent', 'concurrency', undefined);
    });
    expect(screen.getByLabelText('Класс задач для дуэли')).toHaveAttribute(
      'data-disabled',
      'true',
    );
    expect(screen.getByLabelText('Сложность задач для дуэли')).toHaveAttribute(
      'data-disabled',
      'true',
    );
  });
});
