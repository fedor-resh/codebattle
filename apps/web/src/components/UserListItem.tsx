import { Avatar, Button, Group, Paper, Stack, Text } from '@mantine/core';
import { IconSwords } from '@tabler/icons-react';

import { UserStatusBadge, type UserStatus } from './UserStatusBadge';

export type LobbyUser = {
  id: string;
  username: string;
  status: UserStatus;
};

export function UserListItem({
	user,
	isSelf = false,
	inviteDisabled = false,
	onInvite,
}: {
	user: LobbyUser;
	isSelf?: boolean;
	inviteDisabled?: boolean;
	onInvite: (user: LobbyUser) => void;
}) {
	const available = user.status === 'online_available' && !isSelf && !inviteDisabled;

  return (
    <Paper withBorder p="md" radius="md">
      <Group justify="space-between" wrap="nowrap">
        <Group wrap="nowrap" miw={0}>
          <Avatar color="indigo" radius="xl">
            {user.username.slice(0, 2).toUpperCase()}
          </Avatar>
          <Stack gap={4} miw={0}>
            <Text fw={700} truncate>
              {user.username}
            </Text>
            <UserStatusBadge status={user.status} />
          </Stack>
        </Group>
        <Button
          size="xs"
          variant={available ? 'light' : 'default'}
          leftSection={<IconSwords size={15} />}
          disabled={!available}
          onClick={() => onInvite(user)}
          aria-label={`Пригласить ${user.username}`}
        >
			{isSelf ? 'Это вы' : 'Пригласить'}
        </Button>
      </Group>
    </Paper>
  );
}
