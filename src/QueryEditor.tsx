import React, { ChangeEvent, PureComponent } from 'react';
import { Button, InlineField, InlineFieldRow, Input, QueryField, Switch, Select } from '@grafana/ui';
import { MetricFindValue, QueryEditorProps, SelectableValue } from '@grafana/data';
import { CassandraDatasource, CassandraDataSourceOptions } from './datasource';
import { CassandraQuery } from './models';

//const { FormField } = LegacyForms;

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  constructor(props: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>) {
    super(props);

    const { onChange, query } = this.props;
    onChange({ ...query, datasourceId: props.datasource.id });
  }

  getOptions(needType: string): Array<SelectableValue<string>> {
    if (!this.props.query.keyspace || !this.props.query.table) {
      return [];
    }

    this.props.datasource
      .metricFindQuery(this.props.query.keyspace, this.props.query.table)
      .then((columns: MetricFindValue[]) => {
        const columnOptions: Array<SelectableValue<string>> = [];
        columns.forEach((column: MetricFindValue) => {
          if (column.value === needType) {
            columnOptions.push({ label: column.text, value: column.text });
          }
        });

        return columnOptions;
      });

    return [];
  }

  onChangeQueryType = () => {
    const { onChange, query } = this.props;
    onChange({ ...query, rawQuery: !query.rawQuery });
  };

  onQueryTextChange = (request: string) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, target: request });
    onRunQuery();
  };

  onKeyspaceChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, keyspace: event.target.value });
  };

  onTableChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, table: event.target.value });
  };

  onTimeColumnChange = (value: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, columnTime: value.value });
  };

  onValueColumnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, columnValue: event.target.value });
  };

  onIDColumnChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, columnId: event.target.value });
  };

  onIDValueChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, valueId: event.target.value });
  };

  onFilteringChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, filtering: event.target.checked });
  };

  render() {
    const options = this.props;

    return (
      <div>
        <Button icon="pen" variant="secondary" aria-label="Toggle editor mode" onClick={this.onChangeQueryType} />
        {options.query.rawQuery && (
          <QueryField
            placeholder={'Enter a Cassandra query'}
            portalOrigin="cassandra"
            onChange={this.onQueryTextChange}
          />
        )}
        {!options.query.rawQuery && (
          <>
            <InlineFieldRow>
              <InlineField label="Keyspace" tooltip="Specify keyspace to work with" grow>
                <Input
                  name="keyspace"
                  value={this.props.query.keyspace || ''}
                  placeholder="keyspace name"
                  onChange={this.onKeyspaceChange}
                  spellCheck={false}
                  onBlur={this.props.onRunQuery}
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
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Time Column"
                tooltip="Specify name of a timestamp column to identify time (created_at, time etc.)"
                grow
              >
                <Select
                  options={this.getOptions('timestamp')}
                  value={this.props.query.columnTime || ''}
                  onChange={this.onTimeColumnChange}
                  allowCustomValue={true}
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Value Column"
                tooltip="Specify name of a numeric column to retrieve value (temperature, price etc.)"
                grow
              >
                <Input
                  name="value_column"
                  value={this.props.query.columnValue || ''}
                  onChange={this.onValueColumnChange}
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="ID Column"
                tooltip="Specify name of a UUID column to identify the row (id, sensor_id etc.)"
                grow
              >
                <Input
                  name="id_column"
                  value={this.props.query.columnId || ''}
                  onChange={this.onIDColumnChange}
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="ID Value"
                tooltip="Specify UUID value of a column to identify the row (f.e. 123e4567-e89b-12d3-a456-426655440000)"
                grow
              >
                <Input
                  name="value_column"
                  placeholder="123e4567-e89b-12d3-a456-426655440000"
                  value={this.props.query.valueId || '99051fe9-6a9c-46c2-b949-38ef78858dd1'}
                  onChange={this.onIDValueChange}
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Allow filtering"
                tooltip="Allow Filtering can be dangerous practice and we strongly discourage using it"
              >
                <Switch
                  value={this.props.query.filtering}
                  onChange={this.onFilteringChange}
                  onBlur={this.props.onRunQuery}
                />
              </InlineField>
            </InlineFieldRow>
          </>
        )}
      </div>
    );
  }
}
