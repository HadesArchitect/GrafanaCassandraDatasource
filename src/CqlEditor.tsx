import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { getTemplateSrv } from '@grafana/runtime';
import { CodeEditor, CodeEditorSuggestionItem, CodeEditorSuggestionItemKind, Monaco, MonacoEditor } from '@grafana/ui';

/**
 * Time range macros the backend replaces in Query Editor (raw CQL) queries, see pkg/handler/data_query.go.
 * They are not replaced in variable queries.
 */
const TIME_MACROS: CodeEditorSuggestionItem[] = [
  {
    label: '$__timeFrom',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Cassandra macro: start of the time range, Unix epoch milliseconds',
    documentation:
      'Replaced with the start of the panel time range in milliseconds since the Unix epoch, e.g. `registered_at > $__timeFrom`. Works with timestamp columns and with bigint columns holding milliseconds.',
  },
  {
    label: '$__timeTo',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Cassandra macro: end of the time range, Unix epoch milliseconds',
    documentation:
      'Replaced with the end of the panel time range in milliseconds since the Unix epoch, e.g. `registered_at < $__timeTo`. Works with timestamp columns and with bigint columns holding milliseconds.',
  },
  {
    label: '$__unixEpochFrom',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Cassandra macro: start of the time range, Unix epoch seconds',
    documentation:
      'Replaced with the start of the panel time range in seconds since the Unix epoch, for columns that store Unix time in seconds.',
  },
  {
    label: '$__unixEpochTo',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Cassandra macro: end of the time range, Unix epoch seconds',
    documentation:
      'Replaced with the end of the panel time range in seconds since the Unix epoch, for columns that store Unix time in seconds.',
  },
];

/** Grafana global variables, interpolated in the browser for both data and variable queries. */
const GRAFANA_TIME_VARIABLES: CodeEditorSuggestionItem[] = [
  {
    label: '$__from',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Grafana variable: start of the dashboard time range, Unix epoch milliseconds',
    documentation:
      'Grafana global variable with the start of the dashboard time range in milliseconds since the Unix epoch. Supports formats such as `${__from:date:iso}`.',
  },
  {
    label: '$__to',
    kind: CodeEditorSuggestionItemKind.Constant,
    detail: 'Grafana variable: end of the dashboard time range, Unix epoch milliseconds',
    documentation:
      'Grafana global variable with the end of the dashboard time range in milliseconds since the Unix epoch. Supports formats such as `${__to:date:iso}`.',
  },
];

function getDashboardVariableSuggestions(): CodeEditorSuggestionItem[] {
  try {
    return getTemplateSrv()
      .getVariables()
      .map((variable) => ({
        label: `$${variable.name}`,
        kind: CodeEditorSuggestionItemKind.Field,
        detail: 'Dashboard variable',
        documentation: variable.description || variable.label || undefined,
      }));
  } catch (error) {
    // No template service (e.g. outside a dashboard): offer the static suggestions only.
    return [];
  }
}

/**
 * Autocomplete items for the CQL editors: the time range macros (data queries only), Grafana's
 * $__from/$__to and the dashboard variables. Read on every completion request, so variables added
 * while the editor is open are offered too.
 */
export function getCqlSuggestions(includeTimeMacros: boolean): CodeEditorSuggestionItem[] {
  return [...(includeTimeMacros ? TIME_MACROS : []), ...GRAFANA_TIME_VARIABLES, ...getDashboardVariableSuggestions()];
}

// Monaco's default line height in Grafana is 18-19px, so this is about 5 to 20 lines.
const MIN_HEIGHT = 100;
const MAX_HEIGHT = 400;

const clampHeight = (contentHeight: number) => Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, Math.ceil(contentHeight)));

export interface CqlEditorProps {
  /** The committed query text. */
  value: string;
  /** Commits the text. Called when the editor loses focus. */
  onBlur: (value: string) => void;
  /** Called with the current text on Ctrl/Cmd+Enter. */
  onRunQuery: (value: string) => void;
  /** Accessible name of the editor's text area. */
  ariaLabel: string;
  /** Name of the Ctrl/Cmd+Enter action in the editor's command palette. */
  runLabel?: string;
  /** Offer $__timeFrom & co. Only data queries replace them. */
  includeTimeMacros?: boolean;
}

/**
 * A CQL code editor: Grafana's CodeEditor (Monaco) with SQL highlighting, as Monaco has no CQL mode,
 * line numbers, a height that follows the content, autocomplete for macros and variables, and
 * Ctrl/Cmd+Enter bound to `onRunQuery`.
 */
export const CqlEditor = ({
  value,
  onBlur,
  onRunQuery,
  ariaLabel,
  runLabel = 'Run query',
  includeTimeMacros = true,
}: CqlEditorProps) => {
  const [height, setHeight] = useState(MIN_HEIGHT);

  // Monaco keeps the handlers it was given at mount, so they read the latest props through refs.
  const onRunQueryRef = useRef(onRunQuery);
  useEffect(() => {
    onRunQueryRef.current = onRunQuery;
  }, [onRunQuery]);

  // CodeEditor registers one completion provider per editor and replaces it when this function
  // changes, so it must stay the same between renders.
  const getSuggestions = useCallback(() => getCqlSuggestions(includeTimeMacros), [includeTimeMacros]);

  const onEditorDidMount = useCallback(
    (editor: MonacoEditor, monaco: Monaco) => {
      // Actions are scoped to their editor, so with one editor per query Ctrl/Cmd+Enter runs the
      // focused one. Monaco does not dispose actions together with the editor.
      const runAction = editor.addAction({
        id: 'cassandra.runQuery',
        label: runLabel,
        // eslint-disable-next-line no-bitwise
        keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
        run: () => onRunQueryRef.current(editor.getValue()),
      });
      editor.onDidDispose(() => runAction.dispose());

      setHeight(clampHeight(editor.getContentHeight()));
      editor.onDidContentSizeChange((event) => {
        if (event.contentHeightChanged) {
          setHeight(clampHeight(event.contentHeight));
        }
      });
    },
    [runLabel]
  );

  const monacoOptions = useMemo(
    () => ({
      ariaLabel,
      // Keeps the content height independent of the editor height, which follows it.
      scrollBeyondLastLine: false,
      // No horizontal scrollbar to cover the last line.
      wordWrap: 'on' as const,
    }),
    [ariaLabel]
  );

  return (
    <CodeEditor
      value={value}
      language="sql"
      height={height}
      showLineNumbers={true}
      showMiniMap={false}
      onBlur={onBlur}
      getSuggestions={getSuggestions}
      onEditorDidMount={onEditorDidMount}
      monacoOptions={monacoOptions}
    />
  );
};
