import type { PropsWithChildren } from 'react';
import { useState } from 'react';
import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { colorSchemeManager, theme } from './theme';

export function AppProviders({ children }: PropsWithChildren) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            refetchOnWindowFocus: false,
            retry: 1,
            staleTime: 15_000,
          },
        },
      }),
  );

  return (
    <MantineProvider
      theme={theme}
      defaultColorScheme="dark"
      colorSchemeManager={colorSchemeManager}
    >
      <QueryClientProvider client={queryClient}>
        <ModalsProvider>
          <Notifications position="top-right" />
          {children}
        </ModalsProvider>
      </QueryClientProvider>
    </MantineProvider>
  );
}
