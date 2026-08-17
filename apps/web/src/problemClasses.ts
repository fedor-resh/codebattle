import type { ProblemClass } from './api/client';

export const problemClassLabel: Record<ProblemClass, string> = {
  algorithms: 'Алгоритмы',
  concurrency: 'Потоки и горутины',
  oop: 'ООП и ошибки',
};

export const problemClassColor: Record<ProblemClass, string> = {
  algorithms: 'blue',
  concurrency: 'violet',
  oop: 'teal',
};

export const problemClassOptions: Array<{ label: string; value: ProblemClass }> = [
  { label: problemClassLabel.algorithms, value: 'algorithms' },
  { label: problemClassLabel.concurrency, value: 'concurrency' },
  { label: problemClassLabel.oop, value: 'oop' },
];
