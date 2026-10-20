import React, { useEffect, useState } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { Field, LinkButton, TextArea } from '@grafana/ui';
import type { CassandraDatasource } from './datasource';
import { CassandraDataSourceOptions, CassandraQuery, CassandraVariableQuery, normalizeVariableQuery } from './models';

type Props = QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions, CassandraVariableQuery>;

export const VariableQueryEditor = ({ onChange, query }: Props) => {
  // Dashboards saved with plugin 3.x or older store `{ query }` or a bare string.
  const current = normalizeVariableQuery(query);
  const [text, setText] = useState(current.query);

  useEffect(() => {
    setText(current.query);
  }, [current.query]);

  const saveQuery = () => {
    if (text !== current.query) {
      onChange({ ...current, query: text });
    }
  };

  return (
    <>
      <Field
        label="CQL query"
        htmlFor="cassandra-variable-query"
        description="Specify a query that returns variable values and, optionally, their labels. Only strings are allowed here, so use CAST(column as text) if needed. First returned column interpreted as a value and second as a label. Labels should be used in cases when there is an intention to hide exact variable values behind human-readable names in the Grafana UI."
      >
        <TextArea
          id="cassandra-variable-query"
          name="query"
          rows={4}
          onBlur={saveQuery}
          onChange={(event: React.FormEvent<HTMLTextAreaElement>) => setText(event.currentTarget.value)}
          value={text}
          placeholder={'SELECT sensor_id, location FROM sensors.sensors_locations'}
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
