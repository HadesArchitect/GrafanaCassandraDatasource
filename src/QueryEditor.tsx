import React, { ChangeEvent, PureComponent } from 'react';
import { Button, InlineField, InlineFieldRow, Input, QueryField, InlineSwitch, Select } from '@grafana/ui';
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

    const columnOptions: Array<SelectableValue<string>> = [];

    this.props.datasource
      .metricFindQuery(this.props.query.keyspace, this.props.query.table)
      .then((columns: MetricFindValue[]) => {
        columns.forEach((column: MetricFindValue) => {
          if (column.value === needType) {
            columnOptions.push({ label: column.text, value: column.text });
          }
        });

        return columnOptions;
      });

    return columnOptions;
  }

  onChangeQueryType = () => {
    const { onChange, query } = this.props;
    onChange({ ...query, rawQuery: !query.rawQuery });
  };

  onQueryTextChange = (request: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, target: request });
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

  onValueColumnChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, columnValue: event.value });
  };

  onIDColumnChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, columnId: event.value });
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
        {options.query.rawQuery && (
          <InlineFieldRow>
            <InlineField
              label="Cassandra CQL Query"
              labelWidth={30}
              tooltip="Enter Cassandra CQL query. Also you can use $__timeFrom and $__timeTo variables, it will be replaced by chosen range"
              grow
            >
              <QueryField
                placeholder={'Enter a Cassandra query'}
                portalOrigin="cassandra"
                onChange={this.onQueryTextChange}
                onBlur={this.props.onRunQuery}
              />
            </InlineField>
            <Button icon="pen" variant="secondary" aria-label="Toggle editor mode" onClick={this.onChangeQueryType} />
          </InlineFieldRow>
        )}
        {!options.query.rawQuery && (
          <>
            <InlineFieldRow>
              <InlineField label="Keyspace" labelWidth={30} tooltip="Specify keyspace to work with">
                <Input
                  name="keyspace"
                  value={this.props.query.keyspace || ''}
                  placeholder="keyspace name"
                  onChange={this.onKeyspaceChange}
                  spellCheck={false}
                  onBlur={this.props.onRunQuery}
                  width={90}
                />
              </InlineField>
              <Button icon="pen" variant="secondary" aria-label="Toggle editor mode" onClick={this.onChangeQueryType} />
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label="Table" labelWidth={30} tooltip="Specify table to work with">
                <Input
                  name="table"
                  value={this.props.query.table || ''}
                  placeholder="table name"
                  onChange={this.onTableChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Time Column"
                labelWidth={30}
                tooltip="Specify name of a timestamp column to identify time (created_at, time etc.)"
              >
                {/* <Input
                  value={this.props.query.columnTime || ''}
                  placeholder="time column"
                  onChange={this.onTimeColumnChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                /> */}
                <Select
                  allowCustomValue={true}
                  value={this.props.query.columnTime || ''}
                  placeholder="time column"
                  onChange={this.onTimeColumnChange}
                  onBlur={this.props.onRunQuery}
                  options={this.getOptions('timestamp')}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Value Column"
                labelWidth={30}
                tooltip="Specify name of a numeric column to retrieve value (temperature, price etc.)"
              >
                {/* <Input
                  name="value_column"
                  placeholder='value column'
                  value={this.props.query.columnValue || ''}
                  onChange={this.onValueColumnChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                /> */}
                <Select
                  allowCustomValue={true}
                  placeholder="value column"
                  value={this.props.query.columnValue || ''}
                  options={this.getOptions('int')}
                  onChange={this.onValueColumnChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="ID Column"
                labelWidth={30}
                tooltip="Specify name of a UUID column to identify the row (id, sensor_id etc.)"
              >
                {/* <Input
                  name="id_column"
                  placeholder='ID column'
                  value={this.props.query.columnId || ''}
                  onChange={this.onIDColumnChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                /> */}
                <Select
                  allowCustomValue={true}
                  placeholder="ID column"
                  value={this.props.query.columnId || ''}
                  onChange={this.onIDColumnChange}
                  onBlur={this.props.onRunQuery}
                  options={this.getOptions('uuid')}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="ID Value"
                labelWidth={30}
                tooltip="Specify UUID value of a column to identify the row (f.e. 123e4567-e89b-12d3-a456-426655440000)"
              >
                <Input
                  name="value_column"
                  placeholder="123e4567-e89b-12d3-a456-426655440000"
                  value={this.props.query.valueId || ''}
                  onChange={this.onIDValueChange}
                  onBlur={this.props.onRunQuery}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField
                label="Allow filtering"
                labelWidth={30}
                tooltip="Allow Filtering can be dangerous practice and we strongly discourage using it"
              >
                <InlineSwitch
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
