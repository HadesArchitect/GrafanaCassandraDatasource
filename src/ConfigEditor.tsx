import React, { ChangeEvent, PureComponent } from 'react';
import { FieldSet, InlineField, InlineFieldRow, Input, LegacyForms, Select, InlineSwitch, TextArea } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CassandraDataSourceOptions } from './models';

const { SecretFormField } = LegacyForms;

type Props = DataSourcePluginOptionsEditorProps<CassandraDataSourceOptions, Record<string, unknown>>;

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
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

  onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
    });
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


  onCertContentChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        certContent: true,
      },
      secureJsonData: {
        ...options.secureJsonData,
        certContent: event.target.value,
      },
    });
  };

  onRootContentChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        rootContent: true,
      },
      secureJsonData: {
        ...options.secureJsonData,
        rootContent: event.target.value,
      },
    });
  };

  onCaContentChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        caContent: true,
      },
      secureJsonData: {
        ...options.secureJsonData,
        caContent: event.target.value,
      },
    });
  };

  onPasswordReset = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: true,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: event.target.value,
      },
    });
  };

  render() {
    const { onOptionsChange, options } = this.props;
    const { jsonData } = options;

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

    if (!this.props.options.jsonData.consistency || this.props.options.jsonData.consistency === '') {
      this.props.options.jsonData.consistency = consistencyOptions[0].value;
    }

    return (
      <>
        <FieldSet label="Connection settings">
          <InlineFieldRow>
            <InlineField
              label="Host"
              labelWidth={20}
              tooltip="Specify host and port like `192.168.12.134:9042`. You can specify multiple contact points using semicolon, f.e. `host1:9042;host2:9042;host3:9042`"
            >
              <Input
                name="host"
                value={options.url || ''}
                placeholder="cassandra:9042"
                invalid={options.url === ''}
                onChange={this.onHostChange}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Keyspace" labelWidth={20}>
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
            <InlineField label="Consistency" labelWidth={20}>
              <Select
                placeholder="choose consistensy"
                options={consistencyOptions}
                isClearable={false}
                isSearchable={true}
                value={options.jsonData.consistency || consistencyOptions[0]}
                onChange={(value) => {
                  jsonData.consistency = value.value!;
                  onOptionsChange({ ...options, jsonData });
                }}
                width={60}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Credentials"
              tooltip="We strongly recommend to create a custom Cassandra user for Grafana with strictly read-only permissions!"
              labelWidth={20}
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
            <InlineField>
              <SecretFormField
                isConfigured={false}
                value={(options.secureJsonData?.password as string) || ''}
                onReset={this.onPasswordReset}
                onChange={this.onPasswordChange}
                labelWidth={5}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Timeout" labelWidth={20} tooltip="Timeout in seconds. Keep empty for the default value">
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
        </FieldSet>
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
                  <InlineSwitch
                    value={options.jsonData.useCertContent}
                    onChange={this.onUseCertContentChange}
                  />
                </InlineField>
                <InlineField
                  label={options.jsonData.useCertContent ? "Use content" : "Use file paths"}
                  labelWidth={15}
                >
                  <span />
                </InlineField>
              </InlineFieldRow>
            </>
          )}
          {options.jsonData.useCustomTLS && !options.jsonData.useCertContent && (
            <>
              <InlineFieldRow>
                <InlineField
                  label="Certificate Path"
                  labelWidth={30}
                  tooltip="Path to certificate file"
                >
                  <Input
                    value={options.jsonData.certPath}
                    placeholder="certificate path"
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
                <InlineField
                  label="Root Certificate Path"
                  labelWidth={30}
                  tooltip="Path to root certificate file"
                >
                  <Input
                    value={options.jsonData.rootPath}
                    placeholder="root certificate path"
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
                  label="RootCA Certificate Path"
                  labelWidth={30}
                  tooltip="Path to CA certificate file"
                >
                  <Input
                    value={options.jsonData.caPath}
                    placeholder="CA certificate path"
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
                <InlineField
                  label="Certificate Content"
                  labelWidth={30}
                  tooltip="Paste certificate content directly"
                >
                  <TextArea
                    value={(options.secureJsonData?.certContent as string) || ''}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onCertContentChange}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField
                  label="Root Certificate Content"
                  labelWidth={30}
                  tooltip="Paste root certificate content directly"
                >
                  <TextArea
                    value={(options.secureJsonData?.rootContent as string) || ''}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onRootContentChange}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
              <InlineFieldRow>
                <InlineField
                  label="RootCA Certificate Content"
                  labelWidth={30}
                  tooltip="Paste CA certificate content directly"
                >
                  <TextArea
                    value={(options.secureJsonData?.caContent as string) || ''}
                    placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
                    onChange={this.onCaContentChange}
                    rows={6}
                    cols={60}
                  />
                </InlineField>
              </InlineFieldRow>
            </>
          )}
        </FieldSet>
      </>
    );
  }
}
