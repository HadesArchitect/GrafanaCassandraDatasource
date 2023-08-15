import _ from 'lodash';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { DataQueryRequest, DataQueryResponse, DataSourceJsonData, DataSourceInstanceSettings } from '@grafana/data';
import { CassandraQuery } from './models';
import { Observable } from 'rxjs';

export interface CassandraDataSourceOptions extends DataSourceJsonData {
  keyspace: string;
  consistency: string;
  user: string;
  certPath: string;
  rootPath: string;
  caPath: string;
  useCustomTLS: boolean;
  timeout: number;
  allowInsecureTLS: boolean;
}

export class CassandraDatasource extends DataSourceWithBackend<CassandraQuery, CassandraDataSourceOptions> {
  headers: any;
  id: number;

  constructor(instanceSettings: DataSourceInstanceSettings<CassandraDataSourceOptions>) {
    super(instanceSettings);

    this.headers = { 'Content-Type': 'application/json' };

    this.id = instanceSettings.id;

    // annotations default behaviour
    // https://grafana.com/docs/grafana/latest/developers/plugins/create-a-grafana-plugin/extend-a-plugin/add-support-for-annotations/
    this.annotations = {};
  }

  query(options: DataQueryRequest<CassandraQuery>): Observable<DataQueryResponse> {
    if (this.isEditorMode(options)) {
      if (!this.isEditorCompleted(options)) {
        throw new Error('Skipping query execution while not all editor fields are filled');
      }
    } else {
      if (!this.isConfiguratorCompleted(options)) {
        throw new Error('Skipping query execution while not all configurator fields are filled');
      }
    }

    return super.query(this.buildQueryParameters(options));
  }

  isEditorMode(options): boolean {
    return !options.targets[0].rawQuery;
  }

  isEditorCompleted(options): boolean {
    return (
      options.targets[0].keyspace &&
      options.targets[0].table &&
      options.targets[0].columnTime &&
      options.targets[0].columnValue &&
      options.targets[0].columnId &&
      options.targets[0].valueId
    );
  }

  isConfiguratorCompleted(options): boolean {
    return Boolean(options.targets[0].target);
  }

  async getKeyspaces(): Promise<string[]> {
    return this.getResource('keyspaces');
  }

  async getTables(keyspace: string): Promise<string[]> {
    return this.getResource('tables', { keyspace: keyspace });
  }

  async getColumns(keyspace: string, table: string, needType: string): Promise<string[]> {
    return this.getResource('columns', {
      keyspace: keyspace,
      table: table,
      needType: needType,
    });
  }

  buildQueryParameters(options: DataQueryRequest<CassandraQuery>): DataQueryRequest<CassandraQuery> {
    var from = options.range.from.valueOf();
    var to = options.range.to.valueOf();
    options.scopedVars.__timeFrom = { text: from, value: from };
    options.scopedVars.__timeTo = { text: to, value: to };

    //remove placeholder targets
    options.targets = _.filter(options.targets, (target) => {
      return target.target !== 'select metric';
    });

    const targets: CassandraQuery[] = _.map(options.targets, (target) => {
      return {
        datasourceId: target.datasourceId,
        queryType: 'query',

        target: getTemplateSrv().replace(target.target, options.scopedVars),
        refId: target.refId,
        hide: target.hide,
        rawQuery: target.rawQuery,
        filtering: target.filtering,
        keyspace: target.keyspace,
        table: target.table,
        columnTime: target.columnTime,
        columnValue: target.columnValue,
        columnId: target.columnId,
        valueId: target.valueId,
        alias: target.alias,
      };
    });

    options.targets = targets;

    return options;
  }
}
