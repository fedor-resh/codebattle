import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';

import { AppProviders } from './AppProviders';
import { App } from './App';

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
        { id: 'opponent', username: 'go_ninja', status: 'offline' },
      ],
    }),
    heartbeat: vi.fn().mockResolvedValue(undefined),
    getInvitationState: vi.fn().mockResolvedValue({}),
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
});
