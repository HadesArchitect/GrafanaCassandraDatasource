import React from 'react';
import { QueryEditorProps } from '@grafana/data';
import { Field, LinkButton } from '@grafana/ui';
import type { CassandraDatasource } from './datasource';
import { CassandraDataSourceOptions, CassandraQuery, CassandraVariableQuery, normalizeVariableQuery } from './models';
import { CqlEditor } from './CqlEditor';

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions, CassandraVariableQuery>;

export const VariableQueryEditor = ({ onChange, query }: Props) => {
  // Dashboards saved with plugin 3.x or older store `{ query }` or a bare string.
  const current = normalizeVariableQuery(query);

  // Committed on blur and on Ctrl/Cmd+Enter. Grafana then runs the query to refresh the value preview.
  const saveQuery = (text: string) => {
    if (text !== current.query) {
      onChange({ ...current, query: text });
    }
  };

  return (
    <>
      <Field
        label="CQL query"
        description="Specify a query that returns variable values and, optionally, their labels. Only strings are allowed here, so use CAST(column as text) if needed. First returned column interpreted as a value and second as a label. Labels should be used in cases when there is an intention to hide exact variable values behind human-readable names in the Grafana UI. Example: SELECT sensor_id, location FROM sensors.sensors_locations"
      >
        <CqlEditor
          value={current.query}
          ariaLabel="CQL query"
          runLabel="Apply variable query"
          includeTimeMacros={false}
          onBlur={saveQuery}
          onRunQuery={saveQuery}
        />
      </Field>
      <LinkButton
        href="https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md"
        target="_blank"
        rel="noopener noreferrer"
        variant="secondary"
        size="sm"
        icon="external-link-alt"
      >
        Learn more about variables
      </LinkButton>
    </>
  );
};
