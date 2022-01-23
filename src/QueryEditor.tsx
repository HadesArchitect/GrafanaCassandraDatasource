import React, { ChangeEvent, PureComponent } from 'react';
import { Button, InlineField, InlineFieldRow, Input, QueryField, Switch } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { CassandraDatasource, CassandraDataSourceOptions } from './datasource';
import { CassandraQuery } from './models';

//const { FormField } = LegacyForms;

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  constructor(props: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>){
    super(props)

    const { onChange, query } = this.props;
    onChange({ ...query, datasourceId: props.datasource.id });
  }


  onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, target: event.target.value });
    onRunQuery();
  };

  onKeyspaceChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, keyspace: event.target.value });
    onRunQuery();
  };

  onTableChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, table: event.target.value });
    onRunQuery();
  };

  onTimeColumnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, columnTime: event.target.value });
    onRunQuery();
  };

  onValueColumnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, columnValue: event.target.value });
    onRunQuery();
  };

  onIDColumnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, columnId: event.target.value });
    onRunQuery();
  };

  onIDValueChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, valueId: event.target.value });
    onRunQuery();
  };

  onFilteringChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, filtering: event.target.checked });
    onRunQuery();
  };

  render() {
    return (
      
      <>
        <InlineFieldRow>
          <InlineField label="Keyspace" tooltip="Specify keyspace to work with" grow>
            <Input
              name="keyspace"
              value={this.props.query.keyspace || ''}
              placeholder="keyspace name"
              onChange={this.onKeyspaceChange}
              spellCheck={false}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="Table" tooltip="Specify table to work with" grow>
            <Input
              name="table"
              value={this.props.query.table || ''}
              placeholder="table name"
              onChange={this.onTableChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="Time Column" tooltip="Specify name of a timestamp column to identify time (created_at, time etc.)" grow>
            <Input
              name="time_column"
              value={this.props.query.columnTime || ''}
              onChange={this.onTimeColumnChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="Value Column" tooltip="Specify name of a numeric column to retrieve value (temperature, price etc.)" grow>
            <Input
              name="value_column"
              value={this.props.query.columnValue || ''}
              onChange={this.onValueColumnChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="ID Column" tooltip="Specify name of a UUID column to identify the row (id, sensor_id etc.)" grow>
            <Input
              name="id_column"
              value={this.props.query.columnValue || ''}
              onChange={this.onIDColumnChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <InlineField label="ID Value" tooltip="Specify UUID value of a column to identify the row (f.e. 123e4567-e89b-12d3-a456-426655440000)" grow>
            <Input
              name="value_column"
              placeholder="123e4567-e89b-12d3-a456-426655440000"
              value={this.props.query.columnValue || ''}
              onChange={this.onIDValueChange}
            />
          </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
            <InlineField 
              label="Allow filtering"
              tooltip="Allow Filtering can be dangerous practice and we strongly discourage using it"
            >
              <Switch value={this.props.query.filtering} onChange={this.onFilteringChange} />
            </InlineField>
        </InlineFieldRow>
        <InlineFieldRow>
          <Input
            type="hidden"
            name="value_column"
            placeholder="123e4567-e89b-12d3-a456-426655440000"
            value={this.props.query.columnValue || ''}
            onChange={this.onIDValueChange}
          />
        </InlineFieldRow>
        <Button
          icon="pen"
          variant="secondary"
          aria-label="Toggle editor mode"
          onClick={() => {
            console.log("hy")
          }}
        />
        <QueryField
          
          placeholder={'Enter a Graphite query (run with Shift+Enter)'}
          portalOrigin="graphite"
        />
      </>
    );
  }
}
