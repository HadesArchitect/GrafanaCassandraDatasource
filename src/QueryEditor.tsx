import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { CassandraDatasource, CassandraDataSourceOptions } from './datasource';
import { CassandraQuery } from './models';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, target: event.target.value });
  };

  onConstantChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    console.log(event.isTrusted);
    onChange({ ...query });
    // executes the query
    onRunQuery();
  };

  render() {
    return (
      <div className="gf-form">
        <FormField width={4} value={0} onChange={this.onConstantChange} label="Constant" type="number" step="0.1" />
        <FormField
          labelWidth={8}
          value={''}
          onChange={this.onQueryTextChange}
          label="Query Text"
          tooltip="Not used yet"
        />
      </div>
    );
  }
}
