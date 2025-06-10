import * as monaco from 'monaco-editor';
import { listen } from 'vscode-ws-jsonrpc';
import { MonacoLanguageClient, createConnection } from 'monaco-languageclient';
import * as monacoLanguageClient from 'monaco-languageclient';

const uri = monaco.Uri.parse('file:///virtual/main.go');
const model = monaco.editor.createModel('package main\n\nfunc main() {}', 'go', uri);

const editor = monaco.editor.create(document.getElementById('editor'), {
  model,
  language: 'go',
  automaticLayout: true,
});

// WebSocket ke LSP bridge backend
const webSocket = new WebSocket('ws://localhost:8080/ws');

listen({
  webSocket,
  onConnection: (connection) => {
    const languageClient = new MonacoLanguageClient({
      name: 'Gopls Client',
      clientOptions: {
        documentSelector: [{ language: 'go' }],
        initializationOptions: {},
        errorHandler: {
        },
      },
      connectionProvider: {
        get: async () => createConnection(connection),
      },
    });
    languageClient.start();
    connection.onClose(() => languageClient.stop());
  },
});
