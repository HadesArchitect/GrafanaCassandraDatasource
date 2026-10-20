import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { DataQueryRequest, DataSourceInstanceSettings, ScopedVars } from '@grafana/data';
import { lastValueFrom } from 'rxjs';
import { CassandraDatasource } from 'datasource';
import {
  CassandraDataSourceOptions,
  CassandraVariableQuery,
  DEFAULT_VARIABLE_REF_ID,
  normalizeVariableQuery,
} from 'models';
import { CassandraVariableSupport } from 'variables';
import { VariableQueryEditor } from 'VariableQueryEditor';
import { lastCodeEditorProps } from '../testUtils/MockCodeEditor';

// Monaco does not run in jsdom.
jest.mock('@grafana/ui', () => ({
  ...jest.requireActual('@grafana/ui'),
  CodeEditor: jest.requireActual('../testUtils/MockCodeEditor').MockCodeEditor,
}));

const replace = jest.fn((text?: string, scopedVars?: ScopedVars, format?: string) => {
  const zone = scopedVars?.zone?.value;
  return zone !== undefined && text ? text.replace('$zone', `${zone}:${format}`) : text ?? '';
});

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ replace, getVariables: () => [] }),
}));

const makeDatasource = () =>
  new CassandraDatasource({
    id: 1,
    uid: 'cassandra',
    type: 'hadesarchitect-cassandra-datasource',
    name: 'Cassandra',
    access: 'proxy',
    jsonData: {},
    meta: {},
    readOnly: false,
  } as unknown as DataSourceInstanceSettings<CassandraDataSourceOptions>);

const makeRequest = (target: unknown, scopedVars: ScopedVars = {}) =>
  ({ targets: [target], scopedVars } as unknown as DataQueryRequest<CassandraVariableQuery>);

const values = [
  { value: 'a1', text: 'Kitchen' },
  { value: 'a2', text: 'Garage' },
];

describe('normalizeVariableQuery', () => {
  it('keeps a query saved by this version as it is', () => {
    expect(normalizeVariableQuery({ refId: 'A', query: 'SELECT id FROM ks.t' })).toEqual({
      refId: 'A',
      query: 'SELECT id FROM ks.t',
    });
  });

  it('accepts the { query } object saved by plugin 3.x and older', () => {
    expect(normalizeVariableQuery({ query: 'SELECT id FROM ks.t' })).toEqual({
      refId: DEFAULT_VARIABLE_REF_ID,
      query: 'SELECT id FROM ks.t',
    });
  });

  it('accepts a bare query string saved by very old dashboards', () => {
    expect(normalizeVariableQuery('SELECT id FROM ks.t')).toEqual({
      refId: DEFAULT_VARIABLE_REF_ID,
      query: 'SELECT id FROM ks.t',
    });
  });

  it.each([undefined, null, '', {}, { query: 42 }])('turns %p into an empty query', (stored) => {
    expect(normalizeVariableQuery(stored)).toEqual({ refId: DEFAULT_VARIABLE_REF_ID, query: '' });
  });
});

describe('CassandraVariableSupport', () => {
  beforeEach(() => {
    replace.mockClear();
  });

  it('is registered on the data source', () => {
    expect(makeDatasource().variables).toBeInstanceOf(CassandraVariableSupport);
  });

  it.each([
    ['a 4.x query', { refId: 'A', query: 'SELECT id, name FROM ks.t WHERE zone = $zone' }],
    ['a 3.x { query } object', { query: 'SELECT id, name FROM ks.t WHERE zone = $zone' }],
    ['a bare string', 'SELECT id, name FROM ks.t WHERE zone = $zone'],
  ])('runs %s through the variables resource with csv interpolation', async (_, stored) => {
    const ds = makeDatasource();
    const getResource = jest.spyOn(ds, 'getResource').mockResolvedValue(values);
    const scopedVars = { zone: { text: 'eu', value: 'eu' } };

    // Grafana may call `variables.query` detached from the support object.
    const { query } = ds.variables as CassandraVariableSupport;
    const response = await lastValueFrom(query(makeRequest(stored, scopedVars)));

    expect(replace).toHaveBeenCalledWith('SELECT id, name FROM ks.t WHERE zone = $zone', scopedVars, 'csv');
    expect(getResource).toHaveBeenCalledWith('variables', {
      query: 'SELECT id, name FROM ks.t WHERE zone = eu:csv',
    });
    expect(response).toEqual({ data: values });
  });

  it('returns no values when the backend fails', async () => {
    const ds = makeDatasource();
    jest.spyOn(ds, 'getResource').mockRejectedValue(new Error('boom'));
    const warn = jest.spyOn(console, 'warn').mockImplementation(() => {});

    const response = await lastValueFrom((ds.variables as CassandraVariableSupport).query(makeRequest({ query: 'x' })));

    expect(response).toEqual({ data: [] });
    warn.mockRestore();
  });

  it('keeps metricFindQuery working with the legacy { query } shape', async () => {
    const ds = makeDatasource();
    jest.spyOn(ds, 'getResource').mockResolvedValue(values);

    await expect(ds.metricFindQuery({ query: 'SELECT id FROM ks.t' } as CassandraVariableQuery)).resolves.toEqual(
      values
    );
  });
});

describe('VariableQueryEditor', () => {
  const renderEditor = (query: unknown, onChange = jest.fn()) => {
    render(
      <VariableQueryEditor
        datasource={{} as CassandraDatasource}
        query={query as CassandraVariableQuery}
        onChange={onChange}
        onRunQuery={jest.fn()}
      />
    );
    return { onChange, textarea: screen.getByRole('textbox', { name: 'CQL query' }) };
  };

  it.each([
    ['a 4.x query', { refId: 'A', query: 'SELECT id FROM ks.t' }],
    ['a 3.x { query } object', { query: 'SELECT id FROM ks.t' }],
    ['a bare string', 'SELECT id FROM ks.t'],
  ])('shows %s', (_, stored) => {
    const { textarea } = renderEditor(stored);

    expect(textarea).toHaveValue('SELECT id FROM ks.t');
  });

  it('saves an edited legacy query in the current shape on blur', () => {
    const { textarea, onChange } = renderEditor({ query: 'SELECT id FROM ks.t' });

    fireEvent.change(textarea, { target: { value: 'SELECT id, name FROM ks.t' } });
    expect(onChange).not.toHaveBeenCalled();

    fireEvent.blur(textarea);
    expect(onChange).toHaveBeenCalledWith({ refId: DEFAULT_VARIABLE_REF_ID, query: 'SELECT id, name FROM ks.t' });
  });

  it('keeps the refId of a current query', () => {
    const { textarea, onChange } = renderEditor({ refId: 'A', query: '' });

    fireEvent.change(textarea, { target: { value: 'SELECT id FROM ks.t' } });
    fireEvent.blur(textarea);

    expect(onChange).toHaveBeenCalledWith({ refId: 'A', query: 'SELECT id FROM ks.t' });
  });

  it('does not touch the variable when the query did not change', () => {
    const { textarea, onChange } = renderEditor({ query: 'SELECT id FROM ks.t' });

    fireEvent.blur(textarea);

    expect(onChange).not.toHaveBeenCalled();
  });

  it('saves the query on Ctrl+Enter', () => {
    const { textarea, onChange } = renderEditor({ refId: 'A', query: '' });

    fireEvent.change(textarea, { target: { value: 'SELECT id FROM ks.t' } });
    fireEvent.keyDown(textarea, { key: 'Enter', ctrlKey: true });

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith({ refId: 'A', query: 'SELECT id FROM ks.t' });
  });

  it('suggests $__from/$__to but not the data query time macros, which variable queries do not replace', () => {
    renderEditor({ refId: 'A', query: '' });

    const labels = lastCodeEditorProps!.getSuggestions!().map((item) => item.label);

    expect(labels).toEqual(expect.arrayContaining(['$__from', '$__to']));
    expect(labels).not.toContain('$__timeFrom');
    expect(labels).not.toContain('$__unixEpochFrom');
  });
});
