import { DataSourcePlugin } from '@grafana/data';
import { CassandraDataSourceOptions, CassandraDatasource } from './datasource';
import { CassandraQuery } from './models';
import { QueryEditor } from './QueryEditor';
import { ConfigEditor } from './ConfigEditor';

export const plugin = new DataSourcePlugin<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>(
  CassandraDatasource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
