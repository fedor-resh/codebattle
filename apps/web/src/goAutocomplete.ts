import type { editor, IDisposable } from 'monaco-editor/editor/editor.api';

type MonacoAPI = typeof import('monaco-editor/editor/editor.api');

export type GoCompletionTemplate = {
  label: string;
  insertText: string;
  detail: string;
  kind: 'function' | 'keyword' | 'snippet';
};

const snippets: GoCompletionTemplate[] = [
  {
    label: 'if',
    insertText: 'if ${1:condition} {\n\t${0}\n}',
    detail: 'Условие if',
    kind: 'snippet',
  },
  {
    label: 'for',
    insertText: 'for ${1:condition} {\n\t${0}\n}',
    detail: 'Цикл for',
    kind: 'snippet',
  },
  {
    label: 'for range',
    insertText: 'for ${1:index}, ${2:value} := range ${3:values} {\n\t${0}\n}',
    detail: 'Цикл по коллекции',
    kind: 'snippet',
  },
  {
    label: 'switch',
    insertText: 'switch ${1:value} {\ncase ${2:caseValue}:\n\t${0}\ndefault:\n}',
    detail: 'Конструкция switch',
    kind: 'snippet',
  },
  {
    label: 'func',
    insertText: 'func ${1:name}(${2}) ${3:returnType} {\n\t${0}\n}',
    detail: 'Объявление функции',
    kind: 'snippet',
  },
  {
    label: 'defer',
    insertText: 'defer ${1:call}()',
    detail: 'Отложенный вызов',
    kind: 'snippet',
  },
  {
    label: 'import',
    insertText: 'import "${1:package}"',
    detail: 'Импорт пакета',
    kind: 'snippet',
  },
  {
    label: 'import block',
    insertText: 'import (\n\t"${1:package}"\n)',
    detail: 'Блок импортов',
    kind: 'snippet',
  },
];

const builtins: GoCompletionTemplate[] = [
  ['append', 'append(${1:slice}, ${2:value})'],
  ['len', 'len(${1:value})'],
  ['make', 'make(${1:type}, ${2:size})'],
  ['copy', 'copy(${1:destination}, ${2:source})'],
  ['delete', 'delete(${1:map}, ${2:key})'],
  ['panic', 'panic(${1:value})'],
].map(([label, insertText]) => ({
  label,
  insertText,
  detail: 'Встроенная функция Go',
  kind: 'function' as const,
}));

const standardLibrary: Record<string, Array<[string, string]>> = {
  strings: [
    ['Split', 'Split(${1:s}, ${2:sep})'],
    ['Fields', 'Fields(${1:s})'],
    ['Join', 'Join(${1:elems}, ${2:sep})'],
    ['Contains', 'Contains(${1:s}, ${2:substr})'],
    ['HasPrefix', 'HasPrefix(${1:s}, ${2:prefix})'],
    ['HasSuffix', 'HasSuffix(${1:s}, ${2:suffix})'],
    ['TrimSpace', 'TrimSpace(${1:s})'],
    ['ToLower', 'ToLower(${1:s})'],
    ['ToUpper', 'ToUpper(${1:s})'],
    ['ReplaceAll', 'ReplaceAll(${1:s}, ${2:old}, ${3:new})'],
    ['Count', 'Count(${1:s}, ${2:substr})'],
    ['Index', 'Index(${1:s}, ${2:substr})'],
    ['Repeat', 'Repeat(${1:s}, ${2:count})'],
  ],
  strconv: [
    ['Atoi', 'Atoi(${1:s})'],
    ['Itoa', 'Itoa(${1:i})'],
    ['ParseInt', 'ParseInt(${1:s}, ${2:base}, ${3:bitSize})'],
    ['FormatInt', 'FormatInt(${1:i}, ${2:base})'],
    ['ParseFloat', 'ParseFloat(${1:s}, ${2:bitSize})'],
    ['FormatFloat', "FormatFloat(${1:f}, '${2:f}', ${3:-1}, ${4:64})"],
  ],
  sort: [
    ['Ints', 'Ints(${1:values})'],
    ['Strings', 'Strings(${1:values})'],
    ['Slice', 'Slice(${1:values}, func(i, j int) bool {\n\t${0}\n})'],
    ['SearchInts', 'SearchInts(${1:values}, ${2:value})'],
    ['SearchStrings', 'SearchStrings(${1:values}, ${2:value})'],
  ],
  fmt: [
    ['Sprint', 'Sprint(${1:values})'],
    ['Sprintf', 'Sprintf(${1:format}, ${2:values})'],
    ['Println', 'Println(${1:values})'],
    ['Scanf', 'Scanf(${1:format}, ${2:values})'],
  ],
  math: [
    ['Abs', 'Abs(${1:x})'],
    ['Max', 'Max(${1:x}, ${2:y})'],
    ['Min', 'Min(${1:x}, ${2:y})'],
    ['Pow', 'Pow(${1:x}, ${2:y})'],
    ['Sqrt', 'Sqrt(${1:x})'],
    ['Ceil', 'Ceil(${1:x})'],
    ['Floor', 'Floor(${1:x})'],
  ],
};

export function getGoCompletionTemplates(
  linePrefix: string,
  functionSignature: string,
): GoCompletionTemplate[] {
  const memberAccess = linePrefix.match(/\b([A-Za-z_]\w*)\.([A-Za-z_]\w*)?$/);
  if (memberAccess) {
    const packageName = memberAccess[1];
    return (standardLibrary[packageName] ?? []).map(([label, insertText]) => ({
      label,
      insertText,
      detail: `Пакет ${packageName}`,
      kind: 'function',
    }));
  }

  const solve = functionSignature.trim()
    ? [
        {
          label: 'Solve',
          insertText: `${functionSignature.trim()} {\n\t\${0}\n}`,
          detail: `Сигнатура задачи: ${functionSignature.trim()}`,
          kind: 'snippet' as const,
        },
      ]
    : [];

  return [...solve, ...snippets, ...builtins];
}

export function registerGoAutocomplete(
  monaco: MonacoAPI,
  editorInstance: editor.IStandaloneCodeEditor,
  functionSignature: string,
): IDisposable {
  const targetModel = editorInstance.getModel();
  if (!targetModel) return { dispose() {} };

  return monaco.languages.registerCompletionItemProvider('go', {
    triggerCharacters: ['.'],
    provideCompletionItems(model, position) {
      if (model !== targetModel) return { suggestions: [] };

      const word = model.getWordUntilPosition(position);
      const linePrefix = model.getLineContent(position.lineNumber).slice(0, position.column - 1);
      const range = new monaco.Range(
        position.lineNumber,
        word.startColumn,
        position.lineNumber,
        word.endColumn,
      );
      const suggestions = getGoCompletionTemplates(linePrefix, functionSignature).map(
        (template, index) => ({
          label: template.label,
          detail: template.detail,
          kind:
            template.kind === 'snippet'
              ? monaco.languages.CompletionItemKind.Snippet
              : template.kind === 'keyword'
                ? monaco.languages.CompletionItemKind.Keyword
                : monaco.languages.CompletionItemKind.Function,
          insertText: template.insertText,
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          range,
          sortText: `${template.kind === 'snippet' ? '0' : '1'}-${String(index).padStart(3, '0')}`,
        }),
      );

      return { suggestions };
    },
  });
}
