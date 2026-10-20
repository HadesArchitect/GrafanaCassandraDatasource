import { DataSourcePlugin } from '@grafana/data';
import { CassandraDatasource } from './datasource';
import { CassandraQuery, CassandraDataSourceOptions } from './models';
import { QueryEditor } from './QueryEditor';
import { ConfigEditor } from './ConfigEditor';

// The variable query editor is registered by the data source itself, see CassandraVariableSupport.
export const plugin = new DataSourcePlugin<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>(
  CassandraDatasource
)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
