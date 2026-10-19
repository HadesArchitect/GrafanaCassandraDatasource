import React from 'react';
import { render, screen } from '@testing-library/react';
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
