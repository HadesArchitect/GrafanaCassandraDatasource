import _ from 'lodash';
import { DataSourceWithBackend, getBackendSrv, getTemplateSrv, FetchResponse } from '@grafana/runtime';
import {
  toDataFrame,
  DataFrameView,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceJsonData,
  DataSourceInstanceSettings,
  MetricFindValue,
} from '@grafana/data';
import { CassandraQuery } from './models';
import { Observable, lastValueFrom } from 'rxjs';

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

    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }

    this.id = instanceSettings.id;
  }

  query(options: DataQueryRequest<CassandraQuery>): Observable<DataQueryResponse> {
    return super.query(this.buildQueryParameters(options));
  }

  async getKeyspaces(): Promise<MetricFindValue[]> {
    const request: CassandraQuery = {
      datasourceId: this.id,
      queryType: 'keyspaces',
      refId: 'keyspaces',
    };

    var response: FetchResponse<any> = await lastValueFrom(
      getBackendSrv().fetch({
        url: '/api/ds/query',
        method: 'POST',
        data: {
          queries: [request],
        },
      })
    );

    const nameIdx = 0;

    var results: MetricFindValue[] = [];
    response.data.results.keyspaces.frames.forEach((data: { data: any; schema: any }) => {
      new DataFrameView(toDataFrame(data)).forEach((row) => {
        results.push({ text: row[nameIdx], value: row[nameIdx] });
      });
    });

    return new Promise<MetricFindValue[]>((resolve) => {
      resolve(results);
    });
  }

  async getTables(keyspace: string): Promise<MetricFindValue[]> {
    const request: CassandraQuery = {
      datasourceId: this.id,
      queryType: 'tables',
      refId: 'tables',
      keyspace: keyspace,
    };

    var response: FetchResponse<any> = await lastValueFrom(
      getBackendSrv().fetch({
        url: '/api/ds/query',
        method: 'POST',
        data: {
          queries: [request],
        },
      })
    );

    const nameIdx = 0;

    var results: MetricFindValue[] = [];
    response.data.results.tables.frames.forEach((data: { data: any; schema: any }) => {
      new DataFrameView(toDataFrame(data)).forEach((row) => {
        results.push({ text: row[nameIdx], value: row[nameIdx] });
      });
    });

    return new Promise<MetricFindValue[]>((resolve) => {
      resolve(results);
    });
  }

  async metricFindQuery(keyspace: string, table: string): Promise<MetricFindValue[]> {
    const request: CassandraQuery = {
      datasourceId: this.id,
      queryType: 'search',
      refId: 'search',
      keyspace: keyspace,
      table: table,
    };

    var response: FetchResponse<any> = await lastValueFrom(
      getBackendSrv().fetch({
        url: '/api/ds/query',
        method: 'POST',
        data: {
          queries: [request],
        },
      })
    );

    const nameIdx = 0;
    const typeIdx = 1;

    var results: MetricFindValue[] = [];
    response.data.results.search.frames.forEach((data: { data: any; schema: any }) => {
      new DataFrameView(toDataFrame(data)).forEach((row) => {
        results.push({ text: row[nameIdx], value: row[typeIdx] });
      });
    });

    return new Promise<MetricFindValue[]>((resolve) => {
      resolve(results);
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
