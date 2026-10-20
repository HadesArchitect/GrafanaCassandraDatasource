import React, { ChangeEvent, PureComponent } from 'react';
import {
  Alert,
  Field,
  InlineField,
  InlineFieldRow,
  Input,
  InlineSwitch,
  LinkButton,
  RadioButtonGroup,
  Select,
} from '@grafana/ui';
import { CoreApp, QueryEditorProps, SelectableValue } from '@grafana/data';
import { CassandraDatasource } from './datasource';
import { CassandraQuery, CassandraDataSourceOptions, hasTimeFilter, isQueryComplete } from './models';
import { CqlEditor } from './CqlEditor';

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>;

/**
 * Advice shown under a Query Editor query that does not use the dashboard time range. It does not block
 * anything: a query meant to read every row, e.g. for a table, is fine, and the data source's Max rows
 * setting still caps the result.
 */
export const NoTimeFilterWarning = () => (
  <Alert severity="warning" title="This query has no time filter" bottomSpacing={2}>
    It uses none of <code>$__timeFrom</code>, <code>$__timeTo</code>, <code>$__unixEpochFrom</code>,{' '}
    <code>$__unixEpochTo</code>, <code>$__from</code> or <code>$__to</code>, so it reads every matching row whatever the
    dashboard time range, up to the Max rows limit of the data source. To follow the time range, add a condition on the
    time column, for example <code>AND registered_at &gt; $__timeFrom AND registered_at &lt; $__timeTo</code>. Ignore
    this if the query is meant to read all rows.
  </Alert>
);

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

  /**
   * Commits `query` (by default the current one) and runs the panel queries once it is complete.
   * Incomplete queries are only saved: the data source would skip them anyway.
   */
  runQuery = (query: CassandraQuery = this.props.query) => {
    const { app, onChange, onRunQuery } = this.props;
    const queryType: CassandraQuery['queryType'] = app === CoreApp.UnifiedAlerting ? 'alert' : 'query';

    onChange({ ...query, queryType });

    if (isQueryComplete(query)) {
      onRunQuery();
    }
  };

  // For onBlur handlers, which must not pass their event on as the query.
  onBlur = () => {
    this.runQuery();
  };

  onQueryTextCommit = (target: string) => {
    this.runQuery({ ...this.props.query, target });
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
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', margin: '10px 0' }}>
        <RadioButtonGroup
          options={[
            { label: 'Configurator', value: false, icon: 'list-ul' },
            { label: 'Query Editor',  value: true,  icon: 'pen'    },
          ]}
          value={options.query.rawQuery ?? false}
          onChange={(val) => {
            const { onChange, query } = this.props;
            onChange({ ...query, rawQuery: val });
          }}
        />
        <div style={{ display: 'flex', gap: '4px' }}>
          <LinkButton
            href={options.query.rawQuery
              ? 'https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/editor.md'
              : 'https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/configurator.md'}
            target="_blank"
            rel="noreferrer"
            variant="secondary"
            size="sm"
            icon="book"
          >
            Docs
          </LinkButton>
          <LinkButton
            href="https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions"
            target="_blank"
            rel="noreferrer"
            variant="secondary"
            size="sm"
            icon="comment-alt"
          >
            Discussions
          </LinkButton>
        </div>
      </div>
      <div>
        {options.query.rawQuery && (
          <>
            <Field
              label="Cassandra CQL Query"
              description="Limit the query to the dashboard time range with $__timeFrom/$__timeTo (milliseconds) or $__unixEpochFrom/$__unixEpochTo (seconds); $__from/$__to and dashboard variables work too. Type $ for suggestions. Ctrl/Cmd+Enter runs the query."
            >
              <CqlEditor
                value={this.props.query.target ?? ''}
                ariaLabel="Cassandra CQL Query"
                onBlur={this.onQueryTextCommit}
                onRunQuery={this.onQueryTextCommit}
              />
            </Field>
            {Boolean(this.props.query.target?.trim()) && !hasTimeFilter(this.props.query.target) && (
              <NoTimeFilterWarning />
            )}
            <InlineFieldRow>
              <InlineField label="Alias" labelWidth={30} tooltip="Series name override. Plain text or template using column names, e.g. `{{ column1 }}:{{ column2}}`">
                <Input
                    name="alias"
                    onChange={this.onAliasChange}
                    onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
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
                    onBlur={this.onBlur}
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
                  onBlur={this.onBlur}
                />
              </InlineField>
            </InlineFieldRow>
          </>
        )}
      </div>
    </>);
  }
}
