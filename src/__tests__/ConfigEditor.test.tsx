import React from 'react';
import { render } from '@testing-library/react';
import { ConfigEditor } from 'ConfigEditor';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CassandraDataSourceOptions } from 'models';

describe('ConfigEditor', () => {
  it('should render without mutating any props', () => {
    const mockProps: DataSourcePluginOptionsEditorProps<
      CassandraDataSourceOptions,
      Record<string, unknown>
    > = Object.freeze({
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
        }),
        secureJsonData: {},
        secureJsonFields: {},
        readOnly: false,
        withCredentials: false,
      }),
    });

    expect(() => {
      render(<ConfigEditor {...mockProps} />);
    }).not.toThrow();
  });
});
