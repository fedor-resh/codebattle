import { Avatar, Badge, Button, Group, Modal, Stack, Text, Title } from '@mantine/core';
import { IconCheck, IconX } from '@tabler/icons-react';

import type { Invitation } from '../api/client';
import { problemClassLabel } from '../problemClasses';
import { InvitationCountdown } from './InvitationCountdown';

export function InvitationModal({
  invitation,
  loading,
  onAccept,
  onDecline,
}: {
  invitation?: Invitation;
  loading: boolean;
  onAccept: () => void;
  onDecline: () => void;
}) {
  return (
    <Modal
      opened={Boolean(invitation)}
      onClose={onDecline}
      title="Вас вызывают на дуэль"
      centered
      closeOnClickOutside={false}
      closeOnEscape={!loading}
      withCloseButton={!loading}
    >
      {invitation && (
        <Stack>
          <Group wrap="nowrap">
            <Avatar color="indigo" size="lg">
              {invitation.sender.username.slice(0, 2).toUpperCase()}
            </Avatar>
            <div>
              <Title order={4}>{invitation.sender.username}</Title>
              <Text size="sm" c="dimmed">
                предлагает решить задачу на Go
              </Text>
            </div>
          </Group>
          <Badge
            variant="light"
            color={invitation.problem_class === 'concurrency' ? 'violet' : 'blue'}
          >
            {problemClassLabel[invitation.problem_class]}
          </Badge>
          <InvitationCountdown expiresAt={invitation.expires_at} />
          <Group grow>
            <Button
              variant="default"
              leftSection={<IconX size={17} />}
              disabled={loading}
              onClick={onDecline}
            >
              Отклонить
            </Button>
            <Button
              leftSection={<IconCheck size={17} />}
              loading={loading}
              onClick={onAccept}
            >
              Принять
            </Button>
          </Group>
        </Stack>
      )}
    </Modal>
  );
}
