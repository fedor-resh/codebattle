import { Group, Progress, Text } from '@mantine/core';
import { useEffect, useState } from 'react';

function secondsUntil(expiresAt: string) {
  return Math.max(0, Math.ceil((new Date(expiresAt).getTime() - Date.now()) / 1000));
}

export function InvitationCountdown({ expiresAt }: { expiresAt: string }) {
  const [seconds, setSeconds] = useState(() => secondsUntil(expiresAt));

  useEffect(() => {
    setSeconds(secondsUntil(expiresAt));
    const interval = window.setInterval(() => setSeconds(secondsUntil(expiresAt)), 1000);
    return () => window.clearInterval(interval);
  }, [expiresAt]);

  return (
    <div>
      <Group justify="space-between" mb={6}>
        <Text size="xs" c="dimmed">
          Приглашение истечёт
        </Text>
        <Text size="xs" fw={700} aria-live="polite">
          через {seconds} сек.
        </Text>
      </Group>
      <Progress value={(seconds / 30) * 100} color={seconds <= 10 ? 'red' : 'indigo'} />
    </div>
  );
}
