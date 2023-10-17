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
    onChange(state, `${state.rawQuery}`);
  };

  const handleChange = (event: React.FormEvent<HTMLTextAreaElement>) =>
    setState({
      ...state,
      [event.currentTarget.name]: event.currentTarget.value,
    });

  return (
    <>
        <label className="small">Specify a query that returns variable values. If the query returns one column, it will be interpreted as values, if the query returns two columns, the first one is interpreted as a label (human-readable) and the second one as a value. Due to grafana limitations, you MUST cast both the labels and values to strings.</label>
        <br />
        <TextArea
          name="rawQuery"
          className="gf-form-input"
          onBlur={saveQuery}
          onChange={handleChange}
          value={state.rawQuery}
          placeholder={'SELECT CAST(location AS text), CAST(sensor_id AS text) FROM sensors.sensors_locations'}
        />
    </>
  );
};
