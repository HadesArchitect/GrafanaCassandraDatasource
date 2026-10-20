import { DataQuery } from '@grafana/schema';
import { DataSourceJsonData } from '@grafana/data';

export interface CassandraQuery extends DataQuery {
  target?: string;
  queryType: CassandraQueryType;
  filtering?: boolean;
  keyspace?: string;
  table?: string;
  columnTime?: string;
  columnValue?: string;
  columnId?: string;
  valueId?: string;
  rawQuery?: boolean;
  alias?: string;
  instant?: boolean;
}

/**
 * Tells whether a query has everything it needs to be sent to the backend.
 *
 * - Query Editor mode (`rawQuery`): a non-blank CQL statement.
 * - Query Configurator mode: keyspace, table, the time, value and ID columns, and the ID value.
 *
 * Incomplete queries are normal while a query is being written, so callers skip them rather than fail.
 */
export function isQueryComplete(query: CassandraQuery): boolean {
  if (query.rawQuery) {
    return Boolean(query.target?.trim());
  }

  return Boolean(
    query.keyspace && query.table && query.columnTime && query.columnValue && query.columnId && query.valueId
  );
}

/**
 * Placeholders that limit a Query Editor query to the dashboard time range: the plugin's macros, replaced by
 * the backend, and Grafana's global `$__from`/`$__to`, also in their `${__from:...}` forms, replaced in the browser.
 */
export const TIME_FILTER_PLACEHOLDERS = [
  '$__timeFrom',
  '$__timeTo',
  '$__unixEpochFrom',
  '$__unixEpochTo',
  '$__from',
  '$__to',
  '${__from',
  '${__to',
];

/**
 * Tells whether a CQL query uses the dashboard time range. It looks at the query as written, before the
 * browser interpolates `$__from`/`$__to` and dashboard variables, which the backend cannot do. A time range
 * passed in a dashboard variable of its own is not recognised.
 */
export function hasTimeFilter(cql: string | undefined): boolean {
  const text = cql ?? '';
  return TIME_FILTER_PLACEHOLDERS.some((placeholder) => text.includes(placeholder));
}

export interface CassandraVariableQuery extends DataQuery {
  query: string;
}

export const DEFAULT_VARIABLE_REF_ID = 'CassandraVariableQuery';

/**
 * Turns whatever a dashboard has stored as a query variable's query into a CassandraVariableQuery.
 *
 * Grafana stores what the variable editor handed to onChange, so the stored shape depends on the plugin
 * version the variable was last saved with:
 * - 4.x (CustomVariableSupport): `{ refId, query }`
 * - 1.x-3.x (the deprecated setVariableQueryEditor): `{ query }`, without a refId
 * - very old or hand-written dashboards: the bare CQL string
 *
 * All of them keep working without the dashboard being saved again.
 */
export function normalizeVariableQuery(stored: unknown, refId = DEFAULT_VARIABLE_REF_ID): CassandraVariableQuery {
  if (typeof stored === 'string') {
    return { refId, query: stored };
  }

  if (stored && typeof stored === 'object') {
    const { query, ...rest } = stored as Partial<CassandraVariableQuery>;
    return { ...rest, refId: rest.refId || refId, query: typeof query === 'string' ? query : '' };
  }

  return { refId, query: '' };
}

export interface CassandraDataSourceOptions extends DataSourceJsonData {
  keyspace: string;
  consistency: string;
  user: string;
  certPath: string;
  rootPath: string;
  caPath: string;
  useCertContent: boolean;
  useCustomTLS: boolean;
  timeout: number;
  allowInsecureTLS: boolean;
  allowedAuthenticators?: string;
  enableSecureSocksProxy?: boolean;
  /**
   * Largest number of rows a query may return; rows beyond it are dropped with a warning on the panel.
   * Unset uses the default of DEFAULT_MAX_ROWS, 0 means no limit.
   */
  maxRows?: number;
}

/** Row limit applied by the backend when `maxRows` is not set, see pkg/settings.go. */
export const DEFAULT_MAX_ROWS = 100000;

/**
 * Row limit the backend applies to the queries of a data source with these settings, as resolveMaxRows in
 * pkg/settings.go reads them: DEFAULT_MAX_ROWS when `maxRows` is not set, 0 for no limit. Undefined when the
 * value is not a non-negative integer, which the backend rejects.
 */
export function effectiveMaxRows(jsonData?: Partial<CassandraDataSourceOptions>): number | undefined {
  const maxRows: unknown = jsonData?.maxRows;
  if (maxRows === undefined || maxRows === null) {
    return DEFAULT_MAX_ROWS;
  }

  return Number.isInteger(maxRows) && (maxRows as number) >= 0 ? (maxRows as number) : undefined;
}

type CassandraQueryType = 'query' | 'alert';
