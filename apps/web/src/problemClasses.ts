import type { ProblemClass } from './api/client';

export const problemClassLabel: Record<ProblemClass, string> = {
  algorithms: 'Алгоритмы',
  concurrency: 'Потоки и горутины',
};

export const problemClassOptions: Array<{ label: string; value: ProblemClass }> = [
  { label: problemClassLabel.algorithms, value: 'algorithms' },
  { label: problemClassLabel.concurrency, value: 'concurrency' },
];
