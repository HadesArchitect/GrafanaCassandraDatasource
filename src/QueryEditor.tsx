import React, { ChangeEvent, PureComponent, FormEvent } from 'react';
import { Button, InlineField, InlineFieldRow, Input, InlineSwitch, Select, TextArea } from '@grafana/ui';
import { CoreApp, QueryEditorProps, SelectableValue } from '@grafana/data';
import { CassandraDatasource } from './datasource';
import { CassandraQuery, CassandraDataSourceOptions } from './models';

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

function selectable(value?: string): SelectableValue<string> {
  if (!value) {
    return {};
  }

  return { label: value, value: value };
}

export class QueryEditor extends PureComponent<Props> {
  state = {
    keyspaceOptions: [] as Array<SelectableValue<string>>,
    tableOptions: [] as Array<SelectableValue<string>>,
    timeColumnOptions: [] as Array<SelectableValue<string>>,
    valueColumnOptions: [] as Array<SelectableValue<string>>,
    idColumnOptions: [] as Array<SelectableValue<string>>
  };

  constructor(props: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>) {
    super(props);

    const { onChange, query } = this.props;
    onChange({ ...query, datasourceId: props.datasource.id });
  }

  loadKeyspaceOptions = () => {
    this.props.datasource.getKeyspaces().then((keyspaces: string[]) => {
      const keyspaceOptions: Array<SelectableValue<string>> = [];
      keyspaces.forEach((keyspace: string) => {
        keyspaceOptions.push({ label: keyspace, value: keyspace });
      });
      this.setState({ keyspaceOptions });
    }).catch(error => {
      console.warn('QueryEditor: Failed to get keyspaces', error);
      this.setState({ keyspaceOptions: [] });
    });
  };

  loadTableOptions = (keyspace: string) => {
    this.props.datasource.getTables(keyspace).then((tables: string[]) => {
      const tableOptions: Array<SelectableValue<string>> = [];
      tables.forEach((table: string) => {
        tableOptions.push({ label: table, value: table });
      });
      this.setState({ tableOptions });
    }).catch(error => {
      console.warn('QueryEditor: Failed to get tables', error);
      this.setState({ tableOptions: [] });
    });
  };

  loadColumnType = (keyspace: string, table: string, columnType: string, stateKey: string) => {
    this.props.datasource.getColumns(keyspace, table, columnType).then((columns: string[]) => {
      const columnOptions: Array<SelectableValue<string>> = [];
      columns.forEach((column: string) => {
        columnOptions.push({ label: column, value: column });
      });
      this.setState({ [stateKey]: columnOptions });
    }).catch(error => {
      console.warn(`QueryEditor: Failed to get ${columnType} columns`, error);
      this.setState({ [stateKey]: [] });
    });
  };

  loadColumnOptions = (keyspace: string, table: string) => {
    this.loadColumnType(keyspace, table, 'timestamp', 'timeColumnOptions');
    this.loadColumnType(keyspace, table, 'int', 'valueColumnOptions');
    this.loadColumnType(keyspace, table, 'uuid', 'idColumnOptions');
  };

  componentDidMount() {
    // Load keyspace options on initialization.
    this.loadKeyspaceOptions();

    // Get tables and set them as options if keyspace is set.
    if (this.props.query.keyspace) {
      this.loadTableOptions(this.props.query.keyspace);
    }

    // Load column options if both keyspace and table are set.
    if (this.props.query.keyspace && this.props.query.table) {
      this.loadColumnOptions(this.props.query.keyspace, this.props.query.table);
    }
  }

  onRunQuery(
    props: Readonly<Props> &
      Readonly<{
        children?: React.ReactNode;
      }>
  ) {
    this.props.query.queryType = 'query';
    if (this.props.app && this.props.app === CoreApp.UnifiedAlerting) {
      this.props.query.queryType = 'alert';
    }

    const { onChange, query } = this.props;
    onChange({ ...query, queryType: props.query.queryType });

    if ((
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
      ) || (props.query.target && props.query.target !== ''))
    {
      this.props.onRunQuery();
    }
  }


  onChangeQueryType = () => {
    const { onChange, query } = this.props;
    onChange({ ...query, rawQuery: !query.rawQuery });
  };

  onQueryTextChange = (e: FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { onChange, query } = this.props;
    const { value } = e.target as HTMLInputElement | HTMLTextAreaElement;
    onChange({ ...query, target: value });
  };

  onKeyspaceChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    
    // Clear table and column selections when keyspace changes
    onChange({
      ...query,
      keyspace: event.value,
      table: undefined,
      columnTime: undefined,
      columnValue: undefined,
      columnId: undefined
    });

    // Get tables and set them as options when keyspace changes.
    if (event.value) {
      this.loadTableOptions(event.value);
    } else {
      this.setState({ tableOptions: [] });
    }

    // Clear column options
    this.setState({
      timeColumnOptions: [],
      valueColumnOptions: [],
      idColumnOptions: []
    });
  };

  onTableChange = (event: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    
    // Clear column selections when table changes
    onChange({
      ...query,
      table: event.value,
      columnTime: undefined,
      columnValue: undefined,
      columnId: undefined
    });

    // Load column options if both keyspace and table are set
    if (query.keyspace && event.value) {
      this.loadColumnOptions(query.keyspace, event.value);
    } else {
      this.setState({
        timeColumnOptions: [],
        valueColumnOptions: [],
        idColumnOptions: []
      });
    }
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

  onInstantChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, instant: event.target.checked });
  };

  render() {
    const options = this.props;

    return (
      <>
      <Button icon="pen" variant="secondary" onClick={this.onChangeQueryType} style={{ margin: '10px 0' }}>Toggle editor mode</Button>
      <div>
        {options.query.rawQuery && (
          <>
            <InlineFieldRow>
              <InlineField
                label="Cassandra CQL Query"
                labelWidth={30}
                tooltip="Enter Cassandra CQL query. There are $__timeFrom/$__timeTo, $__unixEpochFrom/$__unixEpochTo and $__from/$__to variables to dynamically limit time range in queries. You should always use them to avoid excessive data fetching from DB."
                grow
              >
                <TextArea
                  placeholder={'Enter a CQL query'}
                  onChange={this.onQueryTextChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  value={this.props.query.target}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label="Alias" labelWidth={30} tooltip="Series name override. Plain text or template using column names, e.g. `{{ column1 }}:{{ column2}}`">
                <Input
                    name="alias"
                    onChange={this.onAliasChange}
                    onBlur={() => {
                      this.onRunQuery(this.props);
                    }}
                    value={this.props.query.alias || ''}
                />
              </InlineField>
            </InlineFieldRow>
          </>
        )}
        {!options.query.rawQuery && (
          <>
            <InlineFieldRow>
              <InlineField label="Keyspace" labelWidth={30} tooltip="Specify keyspace to work with">
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.keyspace)}
                  placeholder="keyspace name"
                  onChange={this.onKeyspaceChange}
                  options={this.state.keyspaceOptions}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  width={90}
                />
              </InlineField>
            </InlineFieldRow>
            <InlineFieldRow>
              <InlineField label="Table" labelWidth={30} tooltip="Specify table to work with">
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.table)}
                  placeholder="table name"
                  onChange={this.onTableChange}
                  options={this.state.tableOptions}
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
                <Select
                  allowCustomValue={true}
                  value={selectable(this.props.query.columnTime)}
                  placeholder="time column"
                  onChange={this.onTimeColumnChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  options={this.state.timeColumnOptions}
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
                <Select
                  allowCustomValue={true}
                  placeholder="value column"
                  value={selectable(this.props.query.columnValue)}
                  options={this.state.valueColumnOptions}
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
                tooltip="Specify name of a ID column to identify the row (id, sensor_id etc.)"
              >
                <Select
                  allowCustomValue={true}
                  placeholder="ID column"
                  value={selectable(this.props.query.columnId)}
                  onChange={this.onIDColumnChange}
                  onBlur={() => {
                    this.onRunQuery(this.props);
                  }}
                  options={this.state.idColumnOptions}
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
              <InlineField label="Alias" labelWidth={30} tooltip="Series name override. Plain text or template using column names, e.g. `{{ column1 }}:{{ column2}}`">
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
                  label="Instant"
                  labelWidth={30}
                  tooltip="Queries only first point for each series(PER PARTITION LIMIT 1)"
              >
                <InlineSwitch
                    value={this.props.query.instant}
                    onChange={this.onInstantChange}
                    onBlur={() => {
                      this.onRunQuery(this.props);
                    }}
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
    </>);
  }
}
