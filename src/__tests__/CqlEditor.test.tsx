import React from 'react';
import { render } from '@testing-library/react';
import { CodeEditorSuggestionItemKind } from '@grafana/ui';
import { CqlEditor, getCqlSuggestions } from '../CqlEditor';
import { lastCodeEditorProps, mockMonaco, registeredActions } from '../testUtils/MockCodeEditor';

// Monaco does not run in jsdom.
jest.mock('@grafana/ui', () => ({
  ...jest.requireActual('@grafana/ui'),
  CodeEditor: jest.requireActual('../testUtils/MockCodeEditor').MockCodeEditor,
}));

const mockGetVariables = jest.fn();

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ getVariables: mockGetVariables }),
}));

const TIME_MACROS = ['$__timeFrom', '$__timeTo', '$__unixEpochFrom', '$__unixEpochTo'];

describe('getCqlSuggestions', () => {
  beforeEach(() => {
    mockGetVariables.mockReset().mockReturnValue([]);
  });

  it('offers the time range macros and $__from/$__to for data queries, each with a description', () => {
    const suggestions = getCqlSuggestions(true);

    expect(suggestions.map((item) => item.label)).toEqual([...TIME_MACROS, '$__from', '$__to']);
    for (const item of suggestions) {
      expect(item.kind).toBe(CodeEditorSuggestionItemKind.Constant);
      expect(item.detail).toBeTruthy();
      expect(item.documentation).toBeTruthy();
    }
  });

  it('documents the unit of each macro', () => {
    const byLabel = Object.fromEntries(getCqlSuggestions(true).map((item) => [item.label, item]));

    expect(byLabel.$__timeFrom.detail).toMatch(/milliseconds/);
    expect(byLabel.$__unixEpochTo.detail).toMatch(/seconds/);
  });

  it('leaves the time range macros out for variable queries', () => {
    expect(getCqlSuggestions(false).map((item) => item.label)).toEqual(['$__from', '$__to']);
  });

  it('offers the dashboard variables', () => {
    mockGetVariables.mockReturnValue([
      { name: 'zone', label: 'Zone', description: 'Sensor zone' },
      { name: 'sensor', label: 'Sensor' },
      { name: 'loc' },
    ]);

    const field = CodeEditorSuggestionItemKind.Field;
    expect(getCqlSuggestions(true).slice(-3)).toEqual([
      { label: '$zone', kind: field, detail: 'Dashboard variable', documentation: 'Sensor zone' },
      { label: '$sensor', kind: field, detail: 'Dashboard variable', documentation: 'Sensor' },
      { label: '$loc', kind: field, detail: 'Dashboard variable', documentation: undefined },
    ]);
  });

  it('still offers the built-in suggestions when the template service is not available', () => {
    mockGetVariables.mockImplementation(() => {
      throw new Error('no template service');
    });

    expect(getCqlSuggestions(true)).toHaveLength(6);
  });
});

describe('CqlEditor', () => {
  const renderEditor = (onRunQuery = jest.fn(), onBlur = jest.fn()) => {
    const utils = render(<CqlEditor value="SELECT 1" ariaLabel="CQL" onBlur={onBlur} onRunQuery={onRunQuery} />);
    return { ...utils, onRunQuery, onBlur };
  };

  it('binds Ctrl/Cmd+Enter to a single editor action that runs the current text', () => {
    const { onRunQuery } = renderEditor();

    expect(registeredActions).toHaveLength(1);
    // eslint-disable-next-line no-bitwise
    expect(registeredActions[0].keybindings).toEqual([mockMonaco.KeyMod.CtrlCmd | mockMonaco.KeyCode.Enter]);
    expect(registeredActions[0].label).toBe('Run query');

    registeredActions[0].run({} as never);
    expect(onRunQuery).toHaveBeenCalledWith('SELECT 1');
  });

  it('calls the latest onRunQuery after a re-render, without registering the action again', () => {
    const { rerender, onRunQuery } = renderEditor();
    const next = jest.fn();

    rerender(<CqlEditor value="SELECT 1" ariaLabel="CQL" onBlur={jest.fn()} onRunQuery={next} />);
    registeredActions[0].run({} as never);

    expect(onRunQuery).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith('SELECT 1');
    expect(registeredActions).toHaveLength(1);
  });

  it('starts at the minimum height, commits on blur and sizes the editor to its content', () => {
    const { onBlur } = renderEditor();

    expect(lastCodeEditorProps!.height).toBe(100);
    expect(lastCodeEditorProps!.onBlur).toBe(onBlur);
    expect(lastCodeEditorProps!.monacoOptions).toMatchObject({ ariaLabel: 'CQL', scrollBeyondLastLine: false });
  });
});
