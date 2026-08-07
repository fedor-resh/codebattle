import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';

import { UserStatusBadge } from './UserStatusBadge';

describe('UserStatusBadge', () => {
  it('renders a textual status and accessible label', () => {
    render(
      <MantineProvider>
        <UserStatusBadge status="online_available" />
      </MantineProvider>,
    );

    expect(screen.getByText('Доступен')).toBeInTheDocument();
    expect(screen.getByLabelText('Статус: Доступен')).toBeInTheDocument();
  });
});

