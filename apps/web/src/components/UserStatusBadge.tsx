import { Badge } from '@mantine/core';
import { IconCircleFilled } from '@tabler/icons-react';

export type UserStatus = 'online_available' | 'online_busy' | 'offline';

const statusPresentation: Record<UserStatus, { color: string; label: string }> = {
  online_available: { color: 'green', label: 'Доступен' },
  online_busy: { color: 'yellow', label: 'В матче' },
  offline: { color: 'gray', label: 'Не в сети' },
};

export function UserStatusBadge({ status }: { status: UserStatus }) {
  const presentation = statusPresentation[status];

  return (
    <Badge
      color={presentation.color}
      variant="light"
      leftSection={<IconCircleFilled size={7} aria-hidden="true" />}
      aria-label={`Статус: ${presentation.label}`}
    >
      {presentation.label}
    </Badge>
  );
}

