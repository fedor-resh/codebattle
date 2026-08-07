import { createTheme, localStorageColorSchemeManager } from '@mantine/core';

export const colorSchemeManager = localStorageColorSchemeManager({
  key: 'codebattle-color-scheme',
});

export const theme = createTheme({
  primaryColor: 'indigo',
  defaultRadius: 'md',
  fontFamily:
    'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
  fontFamilyMonospace: '"JetBrains Mono", Consolas, monospace',
  headings: {
    fontFamily:
      'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
  },
  cursorType: 'pointer',
});

