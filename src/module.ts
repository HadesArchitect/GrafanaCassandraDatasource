import { DataSourcePlugin } from '@grafana/data';
import { CassandraDatasource } from './datasource';
import { CassandraQuery,CassandraDataSourceOptions } from './models';
import { QueryEditor } from './QueryEditor';
import { VariableQueryEditor } from './VariableQueryEditor';
import { ConfigEditor } from './ConfigEditor';

export const plugin = new DataSourcePlugin<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>(
  CassandraDatasource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor)
  .setVariableQueryEditor(VariableQueryEditor); // Deprecated, but now documentation on the new approach available atm
