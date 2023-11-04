import { DataQuery } from '@grafana/schema';
import { DataSourceJsonData } from '@grafana/data';

export interface CassandraQuery extends DataQuery {
  target?: string;
  queryType: CassandraQueryType;
  filtering?: boolean;
  keyspace?: string;
  datasourceId?: number;
  table?: string;
  columnTime?: string;
  columnValue?: string;
  columnId?: string;
  valueId?: string;
  rawQuery?: boolean;
  alias?: string;
  instant?: boolean;
}

export interface CassandraVariableQuery {
  query: string;
}

export interface CassandraDataSourceOptions extends DataSourceJsonData {
  keyspace: string;
  consistency: string;
  user: string;
  certPath: string;
  rootPath: string;
  caPath: string;
  useCustomTLS: boolean;
  timeout: number;
  allowInsecureTLS: boolean;
}

type CassandraQueryType = 'query' | 'alert';
