import { DataQuery } from '@grafana/schema';

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

type CassandraQueryType = 'query' | 'alert';
