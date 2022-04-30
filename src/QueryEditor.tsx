import React, { ChangeEvent, PureComponent } from 'react';
import { Button, InlineField, InlineFieldRow, Input, QueryField, InlineSwitch, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { CassandraDatasource, CassandraDataSourceOptions } from './datasource';
import { CassandraQuery } from './models';

//const { FormField } = LegacyForms;

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

function selectable(value?: string): SelectableValue<string> {
  if (!value) {
    return {};
  }

  return { label: value, value: value };
}

export class QueryEditor extends PureComponent<Props> {
  constructor(props: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>) {
    super(props);

    const { onChange, query } = this.props;
    onChange({ ...query, datasourceId: props.datasource.id });
  }

  onRunQuery(
    props: Readonly<Props> &
      Readonly<{
        children?: React.ReactNode;
      }>
  ) {
    if (
      props.query.keyspace &&
      props.query.keyspace !== '' &&
      props.query.table &&
      props.query.table !== '' &&
      props.query.columnTime &&
      props.query.columnTime !== '' &&
      props.query.columnValue &&
      props.query.columnValue !== '' &&
      props.query.columnId &&
      props.query.columnId !== '' &&
      props.query.valueId &&
      props.query.valueId !== ''
    ) {
      this.props.onRunQuery();
    }
  }

  getKeyspaces(): Array<SelectableValue<string>> {
    const result: Array<SelectableValue<string>> = [];
    this.props.datasource.getKeyspaces().then((keyspaces: string[]) => {
      keyspaces.forEach((keyspace: string) => {
        result.push({ label: keyspace, value: keyspace });
      });
    });

    return result;
  }

  getTables(): Array<SelectableValue<string>> {
    if (!this.props.query.keyspace) {
      return [];
    }

    const result: Array<SelectableValue<string>> = [];
    this.props.datasource.getTables(this.props.query.keyspace).then((tables: string[]) => {
      tables.forEach((table: string) => {
        result.push({ label: table, value: table });
      });
    });

    return result;
  }

  getOptions(needType: string): Array<SelectableValue<string>> {
    if (!this.props.query.keyspace || !this.props.query.table) {
      return [];
    }

    const columnOptions: Array<SelectableValue<string>> = [];

    this.props.datasource
      .getColumns(this.props.query.keyspace, this.props.query.table, needType)
      .then((columns: string[]) => {
        columns.forEach((column: string) => {
          columnOptions.push({ label: column, value: column });
        });
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

  onKeyspaceChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, keyspace: event.value });
  };

  onTableChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, table: event.value });
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

  onAliasChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, alias: event.target.value });
  };

  onFilteringChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, filtering: event.target.checked });
  };

  render() {
    const options = this.props;

    this.props.query.queryType = 'query';

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
                placeholder={'Enter a CQL query'}
                portalOrigin="cassandra"
                onChange={this.onQueryTextChange}
                onBlur={this.props.onRunQuery}
                query={this.props.query.target}
              />
            </InlineField>
            <Button icon="pen" variant="secondary" aria-label="Toggle editor mode" onClick={this.onChangeQueryType} />
          </InlineFieldRow>
        )}
        {!options.query.rawQuery && (
          <>
            <InlineFieldRow>
              <InlineField label="Keyspace" labelWidth={30} tooltip="Specify keyspace to work with">
                {/* <Input
                  name="keyspace"
                  value={this.props.query.keyspace || ''}
                  placeholder="keyspace name"
                  onChange={this.onKeyspaceChange}
                  spellCheck={false}
                  onBlur={this.onRunQuery}
                  required
                  width={90}
                /> */}
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.keyspace)}
                  placeholder="keyspace name"
                  onChange={this.onKeyspaceChange}
                  options={this.getKeyspaces()}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  width={90}
                />
              </InlineField>
              <Button icon="pen" variant="secondary" aria-label="Toggle editor mode" onClick={this.onChangeQueryType} />
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label="Table" labelWidth={30} tooltip="Specify table to work with">
                {/* <Input
                  name="table"
                  value={this.props.query.table || ''}
                  placeholder="table name"
                  onChange={this.onTableChange}
                  onBlur={this.onRunQuery}
                  required
                  width={90}
                /> */}
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.table)}
                  placeholder="table name"
                  onChange={this.onTableChange}
                  options={this.getTables()}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
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
                  onBlur={this.onRunQuery}
                  width={90}
                  required
                /> */}
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.columnTime)}
                  placeholder="time column"
                  onChange={this.onTimeColumnChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
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
                  onBlur={this.onRunQuery}
                  width={90}
                  required
                /> */}
                <Select
                  allowCustomValue={true}
                  placeholder="value column"
                  value={selectable(this.props.query.columnValue)}
                  options={this.getOptions('int')}
                  onChange={this.onValueColumnChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
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
                  onBlur={this.onRunQuery}
                  width={90}
                  required
                /> */}
                <Select
                  allowCustomValue={true}
                  placeholder="ID column"
                  value={selectable(this.props.query.columnId)}
                  onChange={this.onIDColumnChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
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
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  width={90}
                  required
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label="Alias" labelWidth={30} tooltip="Alias for graph legend">
                <Input
                  name="alias"
                  placeholder="my alias"
                  value={this.props.query.alias || ''}
                  onChange={this.onAliasChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
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
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                />
              </InlineField>
            </InlineFieldRow>
          </>
        )}
      </div>
    );
  }
}
