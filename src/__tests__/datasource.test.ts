import { DataQueryRequest, DataSourceInstanceSettings, dateTime } from '@grafana/data';
import { BackendSrv, setBackendSrv } from '@grafana/runtime';
import { lastValueFrom, of } from 'rxjs';
import { CassandraDatasource, sortNames } from '../datasource';
import { CassandraDataSourceOptions, CassandraQuery } from '../models';

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getTemplateSrv: () => ({ replace: (text?: string) => (text ?? '').replace('$sensor', 'sensor-1') }),
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

const configurator: CassandraQuery = {
  refId: 'A',
  queryType: 'query',
  keyspace: 'ks',
  table: 't',
  columnTime: 'ts',
  columnValue: 'v',
  columnId: 'id',
  valueId: '$sensor',
};

const raw: CassandraQuery = { refId: 'B', queryType: 'query', rawQuery: true, target: 'SELECT * FROM ks.t' };

const makeRequest = (targets: CassandraQuery[]) =>
  ({
    requestId: 'r1',
    targets,
    scopedVars: {},
    range: { from: dateTime(0), to: dateTime(1000), raw: { from: 'now-1h', to: 'now' } },
    interval: '1s',
    intervalMs: 1000,
    maxDataPoints: 100,
  } as unknown as DataQueryRequest<CassandraQuery>);

describe('CassandraDatasource.filterQuery', () => {
  const ds = makeDatasource();

  it.each<[string, CassandraQuery]>([
    ['a complete configurator query', configurator],
    ['a raw query with CQL', raw],
  ])('keeps %s', (_, query) => {
    expect(ds.filterQuery(query)).toBe(true);
  });

  it.each<[string, CassandraQuery]>([
    ['a new configurator query', { refId: 'A', queryType: 'query' }],
    ['a configurator query without keyspace', { ...configurator, keyspace: '' }],
    ['a configurator query without table', { ...configurator, table: undefined }],
    ['a configurator query without time column', { ...configurator, columnTime: '' }],
    ['a configurator query without value column', { ...configurator, columnValue: '' }],
    ['a configurator query without ID column', { ...configurator, columnId: '' }],
    ['a configurator query without ID value', { ...configurator, valueId: '' }],
    ['a raw query without CQL', { ...raw, target: undefined }],
    ['a raw query with blank CQL', { ...raw, target: '  \n ' }],
    ['a hidden query', { ...raw, hide: true }],
    ['the old "select metric" placeholder', { ...raw, target: 'select metric' }],
  ])('skips %s', (_, query) => {
    expect(ds.filterQuery(query)).toBe(false);
  });

  it('judges each mode by its own fields', () => {
    // CQL left over from Query Editor mode does not complete an incomplete configurator query...
    expect(ds.filterQuery({ ...configurator, valueId: '', target: 'SELECT * FROM ks.t' })).toBe(false);
    // ...and configurator fields do not complete an empty raw query.
    expect(ds.filterQuery({ ...configurator, rawQuery: true, target: '' })).toBe(false);
  });
});

describe('CassandraDatasource.query', () => {
  let fetch: jest.Mock;

  beforeEach(() => {
    fetch = jest.fn(() => of({ data: { results: {} }, status: 200, ok: true }));
    setBackendSrv({ fetch } as unknown as BackendSrv);
  });

  const sentQueries = () => fetch.mock.calls[0][0].data.queries as CassandraQuery[];

  it('sends only the complete queries of a panel, each judged on its own', async () => {
    const ds = makeDatasource();
    const request = makeRequest([
      { refId: 'NEW', queryType: 'query' },
      configurator,
      { ...raw, refId: 'EMPTY', target: '' },
      raw,
      { ...raw, refId: 'HIDDEN', hide: true },
    ]);

    await expect(lastValueFrom(ds.query(request))).resolves.toBeDefined();

    expect(fetch).toHaveBeenCalledTimes(1);
    expect(sentQueries().map((q) => q.refId)).toEqual(['A', 'B']);
    // Variables are interpolated before the query is sent.
    expect(sentQueries()[0].valueId).toBe('sensor-1');
  });

  it('answers with no data, rather than an error, when no query is complete', async () => {
    const ds = makeDatasource();

    const response = await lastValueFrom(ds.query(makeRequest([{ refId: 'A', queryType: 'query' }])));

    expect(response).toEqual({ data: [] });
    expect(fetch).not.toHaveBeenCalled();
  });

  it('does not modify the request it is given', async () => {
    const ds = makeDatasource();
    const targets: CassandraQuery[] = [configurator, { refId: 'NEW', queryType: 'query' }];
    const request = makeRequest(targets);

    await lastValueFrom(ds.query(request));

    expect(request.targets).toBe(targets);
    expect(request.targets).toHaveLength(2);
    expect(request.targets[0].valueId).toBe('$sensor');
  });
});

describe('sortNames', () => {
  it('sorts case-insensitively and numbers by value', () => {
    expect(sortNames(['table10', 'Beta', 'table2', 'alpha', 'system_auth', 'Table1'])).toEqual([
      'alpha',
      'Beta',
      'system_auth',
      'Table1',
      'table2',
      'table10',
    ]);
  });

  it('orders names that differ only in case the same way every time', () => {
    expect(sortNames(['b', 'a', 'B', 'A'])).toEqual(sortNames(['A', 'B', 'a', 'b']));
  });

  it('copies rather than sorts in place, and tolerates a missing list', () => {
    const names = ['b', 'a'];
    expect(sortNames(names)).toEqual(['a', 'b']);
    expect(names).toEqual(['b', 'a']);
    expect(sortNames(null)).toEqual([]);
  });
});

describe('CassandraDatasource dropdown options', () => {
  it('sorts keyspaces, keeps the system ones, and caches the sorted list', async () => {
    const ds = makeDatasource();
    const getResource = jest
      .spyOn(ds, 'getResource')
      .mockResolvedValue(['system_schema', 'Sensors', 'analytics', 'system', 'system_auth']);
    const sorted = ['analytics', 'Sensors', 'system', 'system_auth', 'system_schema'];

    expect(await ds.getKeyspaces()).toEqual(sorted);
    expect(await ds.getKeyspaces()).toEqual(sorted);
    expect(getResource).toHaveBeenCalledTimes(1);
  });

  it('sorts tables per keyspace', async () => {
    const ds = makeDatasource();
    const getResource = jest.spyOn(ds, 'getResource').mockResolvedValue(['t10', 'T1', 't2']);

    expect(await ds.getTables('ks')).toEqual(['T1', 't2', 't10']);
    expect(await ds.getTables('ks')).toEqual(['T1', 't2', 't10']);
    expect(getResource).toHaveBeenCalledTimes(1);
    expect(getResource).toHaveBeenCalledWith('tables', { keyspace: 'ks' });
  });

  it('sorts columns per keyspace, table and type', async () => {
    const ds = makeDatasource();
    const getResource = jest.spyOn(ds, 'getResource').mockResolvedValue(['updated_at', 'Created_at', 'at']);

    expect(await ds.getColumns('ks', 't', 'timestamp')).toEqual(['at', 'Created_at', 'updated_at']);
    expect(getResource).toHaveBeenCalledWith('columns', { keyspace: 'ks', table: 't', needType: 'timestamp' });
  });

  it('returns an empty list when the backend returns none', async () => {
    const ds = makeDatasource();
    jest.spyOn(ds, 'getResource').mockResolvedValue(null);

    expect(await ds.getTables('empty')).toEqual([]);
  });
});
