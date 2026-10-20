import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { ConfigEditor } from 'ConfigEditor';
import { DataSourcePluginOptionsEditorProps, DataSourceSettings } from '@grafana/data';
import { CassandraDataSourceOptions } from 'models';

type Options = DataSourceSettings<CassandraDataSourceOptions, Record<string, unknown>>;

const makeProps = (
  overrides: Partial<Pick<Options, 'secureJsonData' | 'secureJsonFields'>> & {
    jsonData?: Partial<CassandraDataSourceOptions>;
  } = {}
): DataSourcePluginOptionsEditorProps<CassandraDataSourceOptions, Record<string, unknown>> =>
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
      url: 'cassandra:9042',
      user: '',
      database: '',
      basicAuth: false,
      basicAuthUser: '',
      isDefault: false,
      jsonData: Object.freeze({
        keyspace: '',
        consistency: 'ONE',
        user: 'grafana',
        certPath: '',
        rootPath: '',
        caPath: '',
        useCertContent: false,
        useCustomTLS: false,
        timeout: 0,
        allowInsecureTLS: false,
        allowedAuthenticators: '',
        ...overrides.jsonData,
      }),
      secureJsonData: Object.freeze({ ...overrides.secureJsonData }),
      secureJsonFields: Object.freeze({ ...overrides.secureJsonFields }),
      readOnly: false,
      withCredentials: false,
    }),
  });

const lastOptions = (props: ReturnType<typeof makeProps>): Options =>
  (props.onOptionsChange as jest.Mock).mock.calls.at(-1)[0];

const tlsContent = { useCustomTLS: true, useCertContent: true };

describe('ConfigEditor secrets', () => {
  it('lets a password be entered when none is stored', () => {
    const props = makeProps({ secureJsonData: { certContent: 'pem' } });
    render(<ConfigEditor {...props} />);

    const password = screen.getByLabelText('Password');
    expect(password).toHaveAttribute('type', 'password');
    expect(screen.queryByRole('button', { name: 'Reset' })).not.toBeInTheDocument();

    fireEvent.change(password, { target: { value: 's3cret' } });

    const options = lastOptions(props);
    expect(options.secureJsonData).toEqual({ certContent: 'pem', password: 's3cret' });
    // The field must stay editable while typing, so it is not flagged as configured.
    expect(options.secureJsonFields).toEqual({});
  });

  it('hides a stored password and resets it on request', () => {
    const props = makeProps({ secureJsonFields: { password: true, caContent: true } });
    render(<ConfigEditor {...props} />);

    const password = screen.getByLabelText('Password');
    expect(password).toBeDisabled();
    expect(password).toHaveValue('configured');

    fireEvent.click(screen.getByRole('button', { name: 'Reset' }));

    const options = lastOptions(props);
    expect(options.secureJsonFields).toEqual({ password: false, caContent: true });
    expect(options.secureJsonData).toEqual({ password: '' });
  });

  it('lets certificate content be pasted when none is stored', () => {
    const props = makeProps({ jsonData: tlsContent });
    render(<ConfigEditor {...props} />);

    expect(screen.queryByRole('button', { name: 'Reset' })).not.toBeInTheDocument();
    const [certContent, rootContent, caContent] = screen.getAllByPlaceholderText(/BEGIN CERTIFICATE/);

    fireEvent.change(certContent, { target: { value: 'cert' } });
    expect(lastOptions(props).secureJsonData).toEqual({ certContent: 'cert' });

    fireEvent.change(rootContent, { target: { value: 'key' } });
    expect(lastOptions(props).secureJsonData).toEqual({ rootContent: 'key' });

    fireEvent.change(caContent, { target: { value: 'ca' } });
    expect(lastOptions(props).secureJsonData).toEqual({ caContent: 'ca' });
    expect(lastOptions(props).secureJsonFields).toEqual({});
  });

  it('hides stored certificate content and resets each one on its own', () => {
    const props = makeProps({
      jsonData: tlsContent,
      secureJsonFields: { certContent: true, rootContent: true, caContent: true },
    });
    render(<ConfigEditor {...props} />);

    const resets = screen.getAllByRole('button', { name: 'Reset' });
    expect(resets).toHaveLength(3);

    fireEvent.click(resets[1]);

    const options = lastOptions(props);
    expect(options.secureJsonFields).toEqual({ certContent: true, rootContent: false, caContent: true });
    expect(options.secureJsonData).toEqual({ rootContent: '' });
  });
});
