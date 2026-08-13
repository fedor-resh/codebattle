import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';
import { vi } from 'vitest';

import { InvitationModal } from './InvitationModal';

describe('InvitationModal', () => {
  it('shows the problem class before the invitation is accepted', () => {
    render(
      <MantineProvider>
        <InvitationModal
          invitation={{
            id: 'invitation-1',
            sender: { id: 'sender', username: 'go_ninja' },
            receiver: { id: 'receiver', username: 'algo_fox' },
            status: 'pending',
            problem_class: 'concurrency',
            expires_at: new Date(Date.now() + 30_000).toISOString(),
          }}
          loading={false}
          onAccept={vi.fn()}
          onDecline={vi.fn()}
        />
      </MantineProvider>,
    );

    expect(screen.getByText('Потоки и горутины')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Принять' })).toBeEnabled();
  });
});
