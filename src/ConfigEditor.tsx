import React, { ChangeEvent, PureComponent } from 'react';
import { FieldSet, InlineField, InlineFieldRow, Input, LegacyForms, Select } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
//import { MySecureJsonData } from './types';
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

  // Secure field (only sent to the backend)
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
      { label: 'ANY', value: 'ANY' },
      { label: 'ONE', value: 'ONE' },
      { label: 'TWO', value: 'TWO' },
      { label: 'THREE', value: 'THREE' },
      { label: 'QUORUM', value: 'QUORUM' },
      { label: 'ALL', value: 'ALL' },
      { label: 'LOCAL_QUORUM', value: 'LOCAL_QUORUM' },
      { label: 'EACH_QUORUM', value: 'EACH_QUORUM' },
      { label: 'LOCAL_ONE', value: 'LOCAL_ONE' },
    ];

    return (
      <>
        <FieldSet label="Connection settings">
          <InlineFieldRow>
            <InlineField label="Host" tooltip="Specify host and port like `host:9042`" grow>
              <Input
                name="host"
                value={options.url}
                placeholder="cassandra:9042"
                invalid={options.url === ''}
                onChange={this.onHostChange}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Keyspace" grow>
              <Input
                name="keyspace"
                value={options.jsonData.keyspace}
                placeholder="keyspace name"
                invalid={options.jsonData.keyspace === ''}
                onChange={this.onKeyspaceChange}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField label="Consistency" grow>
              <Select
                options={consistencyOptions}
                value={jsonData.consistency || consistencyOptions[0]}
                onChange={(event) => {
                  jsonData.consistency = event.value!;
                  onOptionsChange({ ...options, jsonData });
                }}
              />
            </InlineField>
          </InlineFieldRow>
          <InlineFieldRow>
            <InlineField
              label="Credentials"
              tooltip="We strongly recommend to create a custom Cassandra user with strictly read-only permissions!"
            >
              <Input
                name="user"
                placeholder="user"
                invalid={options.jsonData.user === ''}
                onChange={this.onUserChange}
              />
            </InlineField>
            <InlineField>
              <SecretFormField
                isConfigured={false}
                value={(options.secureJsonData?.password as string) || ''}
                onReset={this.onPasswordReset}
                onChange={this.onPasswordChange}
              />
            </InlineField>
          </InlineFieldRow>
          {/* <InlineFieldRow>
            <InlineField label="Timeout" tooltip="Timeout in seconds. Keep empty for the default value">
              <Input
                name="timeout"
                placeholder=""
                v
                type="number"
                step={1}
                invalid={options.jsonData.keyspace === ''}
                onChange={this.onKeyspaceChange}
              />
            </InlineField>
          </InlineFieldRow> */}
        </FieldSet>
        {/* <FieldSet 
          label="TLS Settings"
        >
          <InlineFieldRow>
            <InlineField 
              label="Custom TLS settings"
              tooltip="Enable if you need custom TLS configuration (usually required using AstraDB, AWS Keyspaces etc.)"
            >
              <InlineSwitch value={options.jsonData} disabled={disabled} onChange={onChange} />
            </InlineField>
          </InlineFieldRow>
        </FieldSet> */}
      </>
    );
  }
}
