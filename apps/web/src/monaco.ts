import { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor/editor/editor.api';
import 'monaco-editor/languages/definitions/go/register';
self.MonacoEnvironment = {
  getWorker() {
    return new Worker(
      new URL('../node_modules/monaco-editor/esm/vs/editor/editor.worker.js', import.meta.url),
      { type: 'module' },
    );
  },
};

loader.config({ monaco });
