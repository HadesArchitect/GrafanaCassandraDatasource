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
  }

  query(options: DataQueryRequest<CassandraQuery>): Observable<DataQueryResponse> {
    return super.query(this.buildQueryParameters(options));
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
