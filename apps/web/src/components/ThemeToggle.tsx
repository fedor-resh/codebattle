import {
  ActionIcon,
  Tooltip,
  useComputedColorScheme,
  useMantineColorScheme,
} from '@mantine/core';
import { IconMoon, IconSun } from '@tabler/icons-react';

export function ThemeToggle() {
  const computedColorScheme = useComputedColorScheme('dark');
  const { setColorScheme } = useMantineColorScheme();
  const isDark = computedColorScheme === 'dark';
  const label = isDark ? 'Включить светлую тему' : 'Включить тёмную тему';

  return (
    <Tooltip label={label}>
      <ActionIcon
        aria-label={label}
        variant="default"
        size="lg"
        onClick={() => setColorScheme(isDark ? 'light' : 'dark')}
      >
        {isDark ? <IconSun size={18} /> : <IconMoon size={18} />}
      </ActionIcon>
    </Tooltip>
  );
}

