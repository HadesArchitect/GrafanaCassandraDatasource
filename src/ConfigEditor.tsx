import React, { ChangeEvent, PureComponent } from 'react';
import {
  FieldSet,
  InlineField,
  InlineFieldRow,
  Input,
  LinkButton,
  Select,
  InlineSwitch,
  SecretInput,
  SecretTextArea,
} from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { config } from '@grafana/runtime';
import { CassandraDataSourceOptions, DEFAULT_MAX_ROWS } from './models';

type Props = DataSourcePluginOptionsEditorProps<CassandraDataSourceOptions, Record<string, unknown>>;

// Keys under secureJsonData, read by the backend from DecryptedSecureJSONData.
type SecretKey = 'password' | 'certContent' | 'rootContent' | 'caContent';

interface State {}

const consistencyOptions = [
  { label: 'ONE', value: 'ONE' },
  { label: 'TWO', value: 'TWO' },
  { label: 'THREE', value: 'THREE' },
  { label: 'QUORUM', value: 'QUORUM' },
  { label: 'ALL', value: 'ALL' },
  { label: 'LOCAL_QUORUM', value: 'LOCAL_QUORUM' },
  { label: 'EACH_QUORUM', value: 'EACH_QUORUM' },
  { label: 'LOCAL_ONE', value: 'LOCAL_ONE' },
];
export class ConfigEditor extends PureComponent<Props, State> {
  componentDidMount() {
    const { onOptionsChange, options } = this.props;
    const { jsonData } = options;

    if (!jsonData.consistency || jsonData.consistency === '') {
      const updatedJsonData = {
        ...jsonData,
        consistency: consistencyOptions[0].value,
      };
      onOptionsChange({ ...options, jsonData: updatedJsonData });
    }
  }

  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    if (event.target.value === '') {
      event.target.setCustomValidity('Cannot be empty');
      event.target.placeholder = 'This field cannot be empty!';
      event.target.style.setProperty('border-color', 'red');
    } else {
      event.target.setCustomValidity('');
      event.target.placeholder = 'cassandra:9042';
      event.target.style.setProperty('border-color', '');
    }
    const { onOptionsChange, options } = this.props;
    const url = event.target.value;
    onOptionsChange({ ...options, url });
  };

  onKeyspaceChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      keyspace: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      user: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onTimeoutChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      timeout: Number(event.target.value),
    };
    onOptionsChange({ ...options, jsonData });
  };

  // An empty field removes the key, so the backend applies its default; 0 is kept, it means no limit.
  onMaxRowsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const value = event.target.value.trim();
    const jsonData = {
      ...options.jsonData,
      maxRows: value === '' ? undefined : Math.trunc(Number(value)),
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAllowedAuthenticatorsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      allowedAuthenticators: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onEnableSecureSocksProxyChange = (event: React.FormEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      enableSecureSocksProxy: event.currentTarget.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUseCustomTLSChange = (event: React.FormEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useCustomTLS: event.currentTarget.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUseCertContentChange = (event: React.FormEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useCertContent: event.currentTarget.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secrets are write-only: Grafana never sends them back, it only reports in
  // secureJsonFields which of them are set. Typing stores the new value in
  // secureJsonData; Reset clears both, so saving removes the stored secret.
  onSecretChange =
    (key: SecretKey) => (event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const { onOptionsChange, options } = this.props;
      onOptionsChange({
        ...options,
        secureJsonData: {
          ...options.secureJsonData,
          [key]: event.currentTarget.value,
        },
      });
    };

  onSecretReset = (key: SecretKey) => () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        [key]: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        [key]: '',
      },
    });
  };

  isSecretConfigured = (key: SecretKey): boolean => Boolean(this.props.options.secureJsonFields?.[key]);

  secretValue = (key: SecretKey): string => (this.props.options.secureJsonData?.[key] as string | undefined) ?? '';

  render() {
    const { onOptionsChange, options } = this.props;
    const { jsonData } = options;

    return (
      <>
        <FieldSet label="Connection settings">
          <InlineFieldRow>
            <InlineField
              label="Host"
              labelWidth={25}
              tooltip="Specify host and port like `192.168.12.134:9042`. You can specify multiple contact points using semicolon, f.e. `host1:9042;host2:9042;host3:9042`"
            >
              <Input
                name="host"
                value={options.url || ''}
                placeholder="cassandra:9042"
                invalid={options.url === ''}
                onChange={this.onHostChange}
                width={60}
                required
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Keyspace" labelWidth={25} tooltip="Optional, defines default keyspace">
              <Input
                name="keyspace"
                value={options.jsonData.keyspace}
                placeholder="keyspace name"
                onChange={this.onKeyspaceChange}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Consistency" labelWidth={25}>
              <Select
                placeholder="choose consistency"
                options={consistencyOptions}
                isClearable={false}
                isSearchable={true}
                value={options.jsonData.consistency || consistencyOptions[0]}
                onChange={(value) => {
                  const updatedJsonData = {
                    ...jsonData,
                    consistency: value.value!,
                  };
                  onOptionsChange({ ...options, jsonData: updatedJsonData });
                }}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Credentials"
              tooltip="We strongly recommend to create a custom Cassandra user for Grafana with strictly read-only permissions!"
              labelWidth={25}
            >
              <Input
                name="user"
                placeholder="user"
                value={options.jsonData.user}
                invalid={options.jsonData.user === ''}
                onChange={this.onUserChange}
                width={25}
              />
            </InlineField>
            <InlineField label="Password" labelWidth={12} htmlFor="cassandra-password">
              <SecretInput
                id="cassandra-password"
                placeholder="password"
                isConfigured={this.isSecretConfigured('password')}
                value={this.secretValue('password')}
                onReset={this.onSecretReset('password')}
                onChange={this.onSecretChange('password')}
                width={25}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Timeout" labelWidth={25} tooltip="Timeout in seconds. Keep empty for the default value">
              <Input
                name="timeout"
                placeholder=""
                type="number"
                step={1}
                value={options.jsonData.timeout}
                onChange={this.onTimeoutChange}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Max rows"
              labelWidth={25}
              htmlFor="cassandra-max-rows"
              invalid={(options.jsonData.maxRows ?? 0) < 0}
              error="Max rows must not be negative"
              tooltip={`Largest number of rows a query may return, across all its series. Reading stops there and the panel shows a warning that the result was truncated. Protects Grafana from running out of memory on a query without a time filter. Leave empty for the default of ${DEFAULT_MAX_ROWS}; 0 means no limit.`}
            >
              <Input
                id="cassandra-max-rows"
                name="maxRows"
                type="number"
                min={0}
                step={1}
                placeholder={String(DEFAULT_MAX_ROWS)}
                value={options.jsonData.maxRows ?? ''}
                onChange={this.onMaxRowsChange}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Allowed authenticators"
              labelWidth={25}
              tooltip='Leave empty unless gocql rejects your cluster with an "unexpected authenticator" error. If it does, add the authenticator class name(s) your cluster announces (e.g. org.apache.cassandra.auth.LDAPAuthenticator), semicolon-separated. Replaces (does not extend) the built-in defaults.'
            >
              <Input
                name="allowedAuthenticators"
                value={options.jsonData.allowedAuthenticators ?? ''}
                placeholder="org.apache.cassandra.auth.PasswordAuthenticator;org.apache.cassandra.auth.LDAPAuthenticator;etc"
                onChange={this.onAllowedAuthenticatorsChange}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
        </FieldSet>
        {config.secureSocksDSProxyEnabled && (
          <FieldSet label="Private Data Source Connect">
            <InlineFieldRow>
              <InlineField
                label="Secure Socks Proxy"
                labelWidth={30}
                htmlFor="enableSecureSocksProxy"
                tooltip="Reach the cluster through Grafana's secure socks proxy instead of connecting to it directly. Required to query a Cassandra cluster that is not exposed to the internet from Grafana Cloud. Use IP addresses as contact points, host names are resolved by the plugin and not by the proxy."
              >
                <InlineSwitch
                  id="enableSecureSocksProxy"
                  value={options.jsonData.enableSecureSocksProxy ?? false}
                  onChange={this.onEnableSecureSocksProxyChange}
                />
              </InlineField>
            </InlineFieldRow>
          </FieldSet>
        )}
        <FieldSet label="TLS Settings">
          <InlineFieldRow>
            <InlineField
              label="Custom TLS settings"
              tooltip="Enable if you need custom TLS configuration (usually required using AstraDB, AWS Keyspaces etc.)"
              labelWidth={30}
            >
              <InlineSwitch value={options.jsonData.useCustomTLS} onChange={this.onUseCustomTLSChange} />
            </InlineField>
          </InlineFieldRow>
          {options.jsonData.useCustomTLS && (
            <>
              <InlineFieldRow>
                <InlineField
                  label="Allow self-signed certificates"
                  labelWidth={30}
                  tooltip="Allow self-signed certificates"
                >
                  <InlineSwitch
                    value={options.jsonData.allowInsecureTLS}
                    onChange={(event: React.FormEvent<HTMLInputElement>) => {
                      const jsonData = {
                        ...options.jsonData,
                        allowInsecureTLS: event.currentTarget.checked,
                      };
                      onOptionsChange({ ...options, jsonData });
                    }}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField
                  label="Certificate input method"
                  labelWidth={30}
                  tooltip="Choose whether to use file paths or paste certificate content directly"
                >
                  <InlineSwitch value={options.jsonData.useCertContent} onChange={this.onUseCertContentChange} />
                </InlineField>
                <InlineField label={options.jsonData.useCertContent ? 'Use content' : 'Use file paths'} labelWidth={15}>
                  <span />
                </InlineField>
              </InlineFieldRow>
            </>
          )}
          {options.jsonData.useCustomTLS && !options.jsonData.useCertContent && (
            <>
              <InlineFieldRow>
                <InlineField label="Public Key Path" labelWidth={30} tooltip="Path to public key file">
                  <Input
                    value={options.jsonData.certPath}
                    placeholder="public key path"
                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                      const jsonData = {
                        ...options.jsonData,
                        certPath: event.currentTarget.value,
                      };
                      onOptionsChange({ ...options, jsonData });
                    }}
                    width={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField label="Private Key Path" labelWidth={30} tooltip="Path to private key file">
                  <Input
                    value={options.jsonData.rootPath}
                    placeholder="Path to private key"
                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                      const jsonData = {
                        ...options.jsonData,
                        rootPath: event.currentTarget.value,
                      };
                      onOptionsChange({ ...options, jsonData });
                    }}
                    width={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField
                  label="Root CA Certificate Path"
                  labelWidth={30}
                  tooltip="Path to Root CA certificate file"
                >
                  <Input
                    value={options.jsonData.caPath}
                    placeholder="Root CA certificate path"
                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                      const jsonData = {
                        ...options.jsonData,
                        caPath: event.currentTarget.value,
                      };
                      onOptionsChange({ ...options, jsonData });
                    }}
                    width={60}
                  />
                </InlineField>
              </InlineFieldRow>
            </>
          )}
          {options.jsonData.useCustomTLS && options.jsonData.useCertContent && (
            <>
              <InlineFieldRow>
                <InlineField label="Public Key Content" labelWidth={30} tooltip="Paste public key directly">
                  <SecretTextArea
                    isConfigured={this.isSecretConfigured('certContent')}
                    value={this.secretValue('certContent')}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onSecretChange('certContent')}
                    onReset={this.onSecretReset('certContent')}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField label="Private Key Content" labelWidth={30} tooltip="Paste private key directly">
                  <SecretTextArea
                    isConfigured={this.isSecretConfigured('rootContent')}
                    value={this.secretValue('rootContent')}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onSecretChange('rootContent')}
                    onReset={this.onSecretReset('rootContent')}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField
                  label="RootCA Certificate Content"
                  labelWidth={30}
                  tooltip="Paste RootCA certificate directly"
                >
                  <SecretTextArea
                    isConfigured={this.isSecretConfigured('caContent')}
                    value={this.secretValue('caContent')}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onSecretChange('caContent')}
                    onReset={this.onSecretReset('caContent')}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
            </>
          )}
        </FieldSet>
        <div style={{ marginTop: '16px', display: 'flex', gap: '8px' }}>
          <LinkButton
            href="https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/index.md"
            target="_blank"
            rel="noreferrer"
            variant="secondary"
            size="sm"
            icon="book"
          >
            Plugin Documentation
          </LinkButton>
          <LinkButton
            href="https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions"
            target="_blank"
            rel="noreferrer"
            variant="secondary"
            size="sm"
            icon="comment-alt"
          >
            GitHub Discussions
          </LinkButton>
        </div>
      </>
    );
  }
}
