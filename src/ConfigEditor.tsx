import React, { ChangeEvent, PureComponent } from 'react';
import { FieldSet, InlineField, InlineFieldRow, Input, LegacyForms, Select, InlineSwitch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { CassandraDataSourceOptions } from './datasource';

const { SecretFormField } = LegacyForms;

type Props = DataSourcePluginOptionsEditorProps<CassandraDataSourceOptions, Record<string, unknown>>;

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
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
            <InlineField label="Host" labelWidth={20} tooltip="Specify host and port like `host:9042`">
              <Input
                name="host"
                value={options.url || 'cassandra:9042'}
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
              tooltip="We strongly recommend to create a custom Cassandra user with strictly read-only permissions!"
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
              <InlineSwitch
                value={options.jsonData.useCustomTLS}
                disabled={false}
                onChange={this.onUseCustomTLSChange}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Allow self-signed certificates"
              labelWidth={30}
              tooltip="Enable if you use self-signed certificates"
            >
              <InlineSwitch
                value={options.jsonData.allowInsecureTLS}
                disabled={false}
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
            <InlineField label="Certificate Path" labelWidth={30}>
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
            <InlineField label="Root Certificate Path" labelWidth={30}>
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
            <InlineField label="RootCA Certificate Path" labelWidth={30}>
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
        </FieldSet>
      </>
    );
  }
}
