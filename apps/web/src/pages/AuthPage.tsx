import {
  Button,
  Container,
  Paper,
  PasswordInput,
  SegmentedControl,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';

import { ApiError, login, register, type User } from '../api/client';

type Mode = 'login' | 'register';

export function AuthPage({ onAuthenticated }: { onAuthenticated: (user: User) => void }) {
  const [mode, setMode] = useState<Mode>('register');
  const form = useForm({
    initialValues: { username: '', password: '' },
    validate: {
      username: (value) =>
        /^[A-Za-z0-9_]{3,24}$/.test(value)
          ? null
          : '3–24 латинских символа, цифры или _',
      password: (value) =>
        value.length >= 6 && value.length <= 128 ? null : 'От 6 до 128 символов',
    },
  });

  const mutation = useMutation({
    mutationFn: (values: typeof form.values) =>
      mode === 'register'
        ? register(values.username, values.password)
        : login(values.username, values.password),
    onSuccess: onAuthenticated,
  });

  const errorMessage =
    mutation.error instanceof ApiError
      ? mutation.error.message
      : mutation.error
        ? 'Сервер недоступен. Проверьте, что API и PostgreSQL запущены.'
        : null;

  return (
    <Container size={440} py={{ base: 48, sm: 80 }}>
      <Stack gap="lg">
        <div>
          <Title ta="center">Начните Go-дуэль</Title>
          <Text c="dimmed" ta="center" mt="xs">
            Для MVP достаточно username и пароля
          </Text>
        </div>

        <Paper withBorder shadow="sm" p="xl" radius="lg">
          <Stack>
            <SegmentedControl
              fullWidth
              value={mode}
              onChange={(value) => {
                setMode(value as Mode);
                mutation.reset();
              }}
              data={[
                { value: 'register', label: 'Регистрация' },
                { value: 'login', label: 'Вход' },
              ]}
            />

            <form onSubmit={form.onSubmit((values) => mutation.mutate(values))}>
              <Stack>
                <TextInput
                  label="Username"
                  placeholder="algo_player"
                  autoComplete="username"
                  required
                  {...form.getInputProps('username')}
                />
                <PasswordInput
                  label="Пароль"
                  placeholder="Минимум 6 символов"
                  autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
                  required
                  {...form.getInputProps('password')}
                />
                {errorMessage && (
                  <Text c="red" size="sm" role="alert">
                    {errorMessage}
                  </Text>
                )}
                <Button type="submit" loading={mutation.isPending}>
                  {mode === 'register' ? 'Создать аккаунт' : 'Войти'}
                </Button>
              </Stack>
            </form>

            <Text size="xs" c="dimmed" ta="center">
              Email и восстановление пароля появятся только при необходимости. Сейчас всё
              минимально.
            </Text>
          </Stack>
        </Paper>
      </Stack>
    </Container>
  );
}
