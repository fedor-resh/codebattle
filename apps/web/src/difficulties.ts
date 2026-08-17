import type { Difficulty } from './api/client';

export type { Difficulty };

export const difficultyLabel: Record<Difficulty, string> = {
  easy: 'Простая',
  medium: 'Средняя',
  hard: 'Сложная',
};

export const difficultyColor: Record<Difficulty, string> = {
  easy: 'green',
  medium: 'yellow',
  hard: 'red',
};

export const difficulties: Difficulty[] = ['easy', 'medium', 'hard'];
