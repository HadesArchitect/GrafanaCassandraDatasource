import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { ConfigEditor } from 'ConfigEditor';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { config } from '@grafana/runtime';
import { CassandraDataSourceOptions } from 'models';

const makeProps = (): DataSourcePluginOptionsEditorProps<CassandraDataSourceOptions, Record<string, unknown>> =>
  Object.freeze({
    onOptionsChange: jest.fn(),
    options: Object.freeze({
      id: 0,
      uid: '',
      orgId: 1,
      name: '',
      typeLogoUrl: '',
      type: '',
      typeName: '',
      access: '',
      url: '',
      user: '',
      database: '',
      basicAuth: false,
      basicAuthUser: '',
      isDefault: false,
      jsonData: Object.freeze({
        keyspace: '',
        consistency: '',
        user: '',
        certPath: '',
        rootPath: '',
        caPath: '',
        useCertContent: false,
        useCustomTLS: false,
        timeout: 0,
        allowInsecureTLS: false,
        allowedAuthenticators: '',
      }),
      secureJsonData: {},
      secureJsonFields: {},
      readOnly: false,
      withCredentials: false,
    }),
  });

describe('ConfigEditor', () => {
  const secureSocksDSProxyEnabled = config.secureSocksDSProxyEnabled;

  afterEach(() => {
    config.secureSocksDSProxyEnabled = secureSocksDSProxyEnabled;
  });

  it('should render without mutating any props', () => {
    expect(() => {
      render(<ConfigEditor {...makeProps()} />);
    }).not.toThrow();
  });

  it('should hide the secure socks proxy toggle when the proxy is not configured on the Grafana instance', () => {
    config.secureSocksDSProxyEnabled = false;

    render(<ConfigEditor {...makeProps()} />);

    expect(screen.queryByLabelText('Secure Socks Proxy')).not.toBeInTheDocument();
  });

  it('should show the secure socks proxy toggle when the proxy is configured on the Grafana instance', () => {
    config.secureSocksDSProxyEnabled = true;

    render(<ConfigEditor {...makeProps()} />);

    expect(screen.getByLabelText('Secure Socks Proxy')).not.toBeChecked();
  });
});

describe('ConfigEditor Max rows', () => {
  const renderWithMaxRows = (maxRows?: number) => {
    const base = makeProps();
    const props = {
      ...base,
      onOptionsChange: jest.fn(),
      options: { ...base.options, jsonData: { ...base.options.jsonData, maxRows } },
    };
    render(<ConfigEditor {...props} />);
    return { props, input: screen.getByLabelText(/Max rows/) };
  };

  const lastMaxRows = (props: { onOptionsChange: jest.Mock }) => {
    const calls = props.onOptionsChange.mock.calls;
    return calls[calls.length - 1][0].jsonData.maxRows;
  };

  it('is empty with the default as placeholder when not set', () => {
    const { input } = renderWithMaxRows(undefined);

    expect(input).toHaveAttribute('type', 'number');
    expect(input).toHaveValue(null);
    expect(input).toHaveAttribute('placeholder', '100000');
  });

  it('shows a stored value, including 0 for no limit', () => {
    expect(renderWithMaxRows(0).input).toHaveValue(0);
  });

  it('stores a number', () => {
    const { props, input } = renderWithMaxRows(undefined);

    fireEvent.change(input, { target: { value: '5000' } });

    expect(lastMaxRows(props)).toBe(5000);
  });

  it('stores 0, which means no limit', () => {
    const { props, input } = renderWithMaxRows(1000);

    fireEvent.change(input, { target: { value: '0' } });

    expect(lastMaxRows(props)).toBe(0);
  });

  it('removes the setting when cleared, so the default applies', () => {
    const { props, input } = renderWithMaxRows(1000);

    fireEvent.change(input, { target: { value: '' } });

    expect(props.onOptionsChange).toHaveBeenCalled();
    expect(lastMaxRows(props)).toBeUndefined();
  });

  it('flags a negative value', () => {
    renderWithMaxRows(-1);

    expect(screen.getByText('Max rows must not be negative')).toBeInTheDocument();
  });
});
