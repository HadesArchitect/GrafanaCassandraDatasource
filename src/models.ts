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
}

type CassandraQueryType = 'query' | 'alert';
