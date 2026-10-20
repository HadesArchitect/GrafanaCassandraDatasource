import React, { useEffect, useRef, useState } from 'react';
import type { CodeEditor, Monaco, MonacoEditor } from '@grafana/ui';

/**
 * Stand-in for @grafana/ui's CodeEditor in Jest: Monaco does not run in jsdom. It renders a textarea
 * labelled with `monacoOptions.ariaLabel` and fakes the small part of the Monaco API the plugin uses,
 * so Ctrl/Cmd+Enter on the textarea triggers the actions registered through onEditorDidMount.
 *
 * Use it with:
 *   jest.mock('@grafana/ui', () => ({
 *     ...jest.requireActual('@grafana/ui'),
 *     CodeEditor: jest.requireActual('../testUtils/MockCodeEditor').MockCodeEditor,
 *   }));
 */

type CodeEditorProps = React.ComponentProps<typeof CodeEditor>;

interface FakeAction {
  id: string;
  label: string;
  keybindings?: number[];
  run: (editor: MonacoEditor) => void;
}

// Same values as monaco.KeyMod.CtrlCmd and monaco.KeyCode.Enter.
export const mockMonaco = {
  KeyMod: { CtrlCmd: 2048, Shift: 1024, Alt: 512, WinCtrl: 256 },
  KeyCode: { Enter: 3 },
} as unknown as Monaco;

/** The props of the most recently rendered editor, to inspect what the plugin passed. */
export let lastCodeEditorProps: CodeEditorProps | undefined;
/** The actions registered on the most recently mounted editor. */
export let registeredActions: FakeAction[] = [];

export const MockCodeEditor = (props: CodeEditorProps) => {
  lastCodeEditorProps = props;
  const [text, setText] = useState(props.value);
  // Kept in step with `text` by the handlers below, not during render.
  const textRef = useRef(text);
  const actions = useRef<FakeAction[]>([]);

  useEffect(() => {
    textRef.current = props.value;
    setText(props.value);
  }, [props.value]);

  useEffect(() => {
    const noop = { dispose: () => {} };
    const editor = {
      getValue: () => textRef.current,
      getContentHeight: () => 40,
      onDidContentSizeChange: () => noop,
      onDidDispose: () => noop,
      addAction: (action: FakeAction) => {
        actions.current.push(action);
        registeredActions = actions.current;
        return noop;
      },
    } as unknown as MonacoEditor;
    actions.current = [];
    registeredActions = actions.current;
    props.onEditorDidMount?.(editor, mockMonaco);

    // Only Monaco's mount matters, like the real editor which calls this once.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key !== 'Enter' || !(event.ctrlKey || event.metaKey)) {
      return;
    }
    // eslint-disable-next-line no-bitwise
    const ctrlEnter = mockMonaco.KeyMod.CtrlCmd | mockMonaco.KeyCode.Enter;
    actions.current
      .filter((action) => action.keybindings?.includes(ctrlEnter))
      .forEach((action) => action.run({ getValue: () => textRef.current } as unknown as MonacoEditor));
  };

  return (
    <textarea
      data-testid="mock-code-editor"
      data-language={props.language}
      aria-label={props.monacoOptions?.ariaLabel}
      value={text}
      onChange={(event) => {
        textRef.current = event.target.value;
        setText(event.target.value);
      }}
      onBlur={() => props.onBlur?.(textRef.current)}
      onKeyDown={onKeyDown}
    />
  );
};
