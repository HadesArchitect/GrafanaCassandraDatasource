import React, { useState } from 'react';
import { CassandraVariableQuery } from './models';
import { TextArea } from '@grafana/ui';

interface VariableQueryProps {
  query: CassandraVariableQuery;
  onChange: (query: CassandraVariableQuery, definition: string) => void;
}

export const VariableQueryEditor = ({ onChange, query }: VariableQueryProps) => {
  const [state, setState] = useState(query);

  const saveQuery = () => {
    onChange(state, `${state.query}`);
  };

  const handleChange = (event: React.FormEvent<HTMLTextAreaElement>) =>
    setState({
      ...state,
      [event.currentTarget.name]: event.currentTarget.value,
    });

  return (
    <>
        <label className="small">Specify a query that returns variable values and, optionally, their labels. Only strings are allowed here, so use CAST(column as text) if needed. First returned column interpreted as a value and second as a label. Labels should be used in cases when there is an intention to hide exact variable values behind human-readable names in the Grafana UI.</label>
        <br />
        <TextArea
          name="query"
          className="gf-form-input"
          onBlur={saveQuery}
          onChange={handleChange}
          value={state.query}
          placeholder={'SELECT sensor_id, location FROM sensors.sensors_locations'}
        />
    </>
  );
};
