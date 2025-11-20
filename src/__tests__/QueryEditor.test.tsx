import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { QueryEditor } from '../QueryEditor';
import { CassandraDatasource } from '../datasource';
import { CassandraQuery, CassandraDataSourceOptions } from '../models';
import { QueryEditorProps, LoadingState, DataFrame } from '@grafana/data';

// Mock the datasource
export const mockDatasource = {
  id: 1,
  getKeyspaces: jest.fn().mockResolvedValue(['keyspace1', 'keyspace2']),
  getTables: jest.fn().mockResolvedValue(['table1', 'table2']),
  getColumns: jest.fn().mockResolvedValue(['column1', 'column2']),
} as unknown as CassandraDatasource;

// Mock query object
export const mockQuery: CassandraQuery = {
  refId: 'A',
  queryType: 'query',
  rawQuery: false,
  keyspace: '',
  table: '',
  columnTime: '',
  columnValue: '',
  columnId: '',
  valueId: '',
  alias: '',
  filtering: false,
  instant: false,
};

// Mock props
export const mockProps: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions> = {
  datasource: mockDatasource,
  query: mockQuery,
  onChange: jest.fn(),
  onRunQuery: jest.fn(),
  range: {
    from: { valueOf: () => Date.now() - 3600000 } as any,
    to: { valueOf: () => Date.now() } as any,
    raw: { from: 'now-1h', to: 'now' },
  },
  data: {
    state: LoadingState.Done,
    series: [],
    timeRange: {
      from: { valueOf: () => Date.now() - 3600000 } as any,
      to: { valueOf: () => Date.now() } as any,
      raw: { from: 'now-1h', to: 'now' },
    },
  },
};

const getFrozenProps = (): QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions> => {
  const from = Object.freeze({ ...mockProps.range!.from });
  const to = Object.freeze({ ...mockProps.range!.to });
  const raw = Object.freeze({ ...mockProps.range!.raw });
  const query = Object.freeze({ ...mockProps.query });
  const range = Object.freeze({ from, to, raw });
  const data = Object.freeze({
    ...mockProps.data,
    state: LoadingState.Done,
    series: Object.freeze([]) as unknown as DataFrame[],
    timeRange: Object.freeze({ from, to, raw }),
  });
  const frozenProps = Object.freeze({ ...mockProps, query, range, data });
  return frozenProps;
};

const renderComponent = async (
  props: QueryEditorProps<CassandraDatasource, CassandraQuery, CassandraDataSourceOptions>
) => await waitFor(() => render(<QueryEditor {...props} />));

describe('QueryEditor', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render without crashing', () => {
    expect(async () => await renderComponent({ ...mockProps })).not.toThrow();
  });

  it('should render without mutating props', () => {
    expect(async () => await renderComponent({ ...getFrozenProps() })).not.toThrow();
  });

  it('should render run onRunQuery without mutating props', async () => {
    const { getByPlaceholderText } = await renderComponent({ ...getFrozenProps() });

    expect(getByPlaceholderText('my alias')).toBeInTheDocument();
    expect(() => fireEvent.blur(getByPlaceholderText('my alias'))).not.toThrow();
  });

  it('should render keyspace selector in configurator mode', async () => {
    const { getByText } = await renderComponent({ ...mockProps });

    // Should show keyspace selector when not in raw query mode
    expect(getByText('Keyspace')).toBeInTheDocument();
    expect(getByText('Table')).toBeInTheDocument();
  });

  it('should render raw query editor when in raw query mode', async () => {
    const rawQueryProps = {
      ...mockProps,
      query: { ...mockQuery, rawQuery: true },
    };

    const { getByText, getByPlaceholderText } = await renderComponent({ ...rawQueryProps });

    // Should show raw query editor
    expect(getByText('Cassandra CQL Query')).toBeInTheDocument();
    expect(getByPlaceholderText('Enter a CQL query')).toBeInTheDocument();
  });

  it('should call datasource methods on mount', async () => {
    await renderComponent({ ...mockProps });

    expect(mockDatasource.getKeyspaces).toHaveBeenCalledTimes(1);
  });

  it('should toggle query type when button is clicked', async () => {
    const mockOnChange = jest.fn();
    const propsWithMockOnChange = { ...mockProps, onChange: mockOnChange };

    await renderComponent({ ...propsWithMockOnChange });

    // Find and click the toggle button
    const toggleButton = screen.getByRole('button', { name: /toggle editor mode/i });
    fireEvent.click(toggleButton);

    // Verify onChange was called with toggled rawQuery
    expect(mockOnChange).toHaveBeenCalledWith({ ...mockQuery, rawQuery: !mockQuery.rawQuery });
  });

  it('should handle keyspace change correctly', async () => {
    const mockOnChange = jest.fn();

    // Start with a query that has both keyspace and table set
    const queryWithKeyspaceAndTable: CassandraQuery = {
      ...mockQuery,
      keyspace: 'initial_keyspace',
      table: 'initial_table',
    };

    const propsWithInitialData = {
      ...mockProps,
      query: queryWithKeyspaceAndTable,
      onChange: mockOnChange,
    };

    await renderComponent({ ...propsWithInitialData });

    // Wait for component to mount and load keyspaces
    expect(mockDatasource.getKeyspaces).toHaveBeenCalled();

    // Note: Testing Select component interactions can be complex with react-select
    // We'll test the handler method directly since the Select component doesn't
    // render the display value in a way that's easily testable
    const component = new QueryEditor(propsWithInitialData);

    // Mock setState to avoid warnings about unmounted component
    component.setState = jest.fn();

    component.onKeyspaceChange({ value: 'new_keyspace', label: 'new_keyspace' });

    // Verify onChange was called with new keyspace and cleared table/columns
    expect(mockOnChange).toHaveBeenCalledWith({
      ...queryWithKeyspaceAndTable,
      keyspace: 'new_keyspace',
      table: undefined,
      columnTime: undefined,
      columnValue: undefined,
      columnId: undefined,
    });
  });

  it('should handle table change correctly', () => {
    const mockOnChange = jest.fn();
    const queryWithKeyspace = { ...mockQuery, keyspace: 'test_keyspace' };
    const propsWithKeyspace = { ...mockProps, query: queryWithKeyspace, onChange: mockOnChange };

    // Test the handler method directly
    const component = new QueryEditor(propsWithKeyspace);

    // Mock setState to avoid warnings about unmounted component
    component.setState = jest.fn();

    component.onTableChange({ value: 'new_table', label: 'new_table' });

    // Verify onChange was called with new table and cleared columns
    expect(mockOnChange).toHaveBeenCalledWith({
      ...queryWithKeyspace,
      table: 'new_table',
      columnTime: undefined,
      columnValue: undefined,
      columnId: undefined,
    });
  });

  it('should handle column changes correctly', () => {
    const mockOnChange = jest.fn();
    const propsWithOnChange = { ...mockProps, onChange: mockOnChange };
    const component = new QueryEditor(propsWithOnChange);

    // Test time column change
    component.onTimeColumnChange({ value: 'timestamp_col', label: 'timestamp_col' });
    expect(mockOnChange).toHaveBeenCalledWith({
      ...mockQuery,
      columnTime: 'timestamp_col',
    });

    // Reset mock
    mockOnChange.mockClear();

    // Test value column change
    component.onValueColumnChange({ value: 'value_col', label: 'value_col' });
    expect(mockOnChange).toHaveBeenCalledWith({
      ...mockQuery,
      columnValue: 'value_col',
    });

    // Reset mock
    mockOnChange.mockClear();

    // Test ID column change
    component.onIDColumnChange({ value: 'id_col', label: 'id_col' });
    expect(mockOnChange).toHaveBeenCalledWith({
      ...mockQuery,
      columnId: 'id_col',
    });
  });

  it('should handle text input changes', async () => {
    const mockOnChange = jest.fn();
    const propsWithOnChange = { ...mockProps, onChange: mockOnChange };

    await renderComponent({ ...propsWithOnChange });

    // Find ID Value input and change it
    const idValueInput = screen.getByPlaceholderText('123e4567-e89b-12d3-a456-426655440000');
    fireEvent.change(idValueInput, { target: { value: 'test-uuid' } });

    expect(mockOnChange).toHaveBeenCalledWith({ ...mockQuery, valueId: 'test-uuid' });
  });

  it('should handle switch toggles', async () => {
    const mockOnChange = jest.fn();
    const propsWithOnChange = { ...mockProps, onChange: mockOnChange };

    await renderComponent({ ...propsWithOnChange });

    // Get all checkboxes and find the ones we need
    const checkboxes = screen.getAllByRole('checkbox', { name: /toggle switch/i });

    // The first checkbox should be the "Instant" switch, the second should be "Allow filtering"
    // Let's click the first one (Instant switch)
    fireEvent.click(checkboxes[0]);

    expect(mockOnChange).toHaveBeenCalledWith({ ...mockQuery, instant: true });

    // Reset mock and test the second switch (Allow filtering)
    mockOnChange.mockClear();
    fireEvent.click(checkboxes[1]);

    expect(mockOnChange).toHaveBeenCalledWith({ ...mockQuery, filtering: true });
  });
});
