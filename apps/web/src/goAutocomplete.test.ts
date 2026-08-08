import { describe, expect, it } from 'vitest';

import { getGoCompletionTemplates } from './goAutocomplete';

describe('getGoCompletionTemplates', () => {
  it('uses the function signature of the current problem', () => {
    const suggestions = getGoCompletionTemplates('Sol', 'func Solve(input string) string');
    const solve = suggestions.find((suggestion) => suggestion.label === 'Solve');

    expect(solve?.insertText).toBe('func Solve(input string) string {\n\t${0}\n}');
    expect(suggestions.map((suggestion) => suggestion.label)).toContain('for range');
    expect(suggestions.map((suggestion) => suggestion.label)).toContain('append');
  });

  it('returns package members after strings dot', () => {
    const suggestions = getGoCompletionTemplates('value := strings.Spl', 'func Solve() string');
    const labels = suggestions.map((suggestion) => suggestion.label);

    expect(labels).toEqual(expect.arrayContaining(['Split', 'Fields', 'Join']));
    expect(labels).not.toContain('Solve');
  });

  it('does not offer unrelated completions for an unknown package', () => {
    expect(getGoCompletionTemplates('custom.', 'func Solve() string')).toEqual([]);
  });
});
