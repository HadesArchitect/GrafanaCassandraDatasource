import _ from 'lodash';
import {
  DataSourceWithBackend,
  getBackendSrv,
  getTemplateSrv,
} from '@grafana/runtime';
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

//import { DataFrame } from '@grafana/data';

export interface CassandraDataSourceOptions extends DataSourceJsonData {
  keyspace: string;
  consistency: string;
  user: string;
  certPath: string;
  rootPath: string;
  caPath: string;
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
    return super.query(this.buildQueryParameters(options))
  }
  
  async metricFindQuery(keyspace: string, table: string): Promise<MetricFindValue[]> {  
    const request: CassandraQuery = {
      datasourceId: this.id,
      queryType: "search",
      refId: "search",
      keyspace: keyspace,
      table: table
    };
   
    var response = await lastValueFrom(getBackendSrv().fetch({
      url: '/api/ds/query',
      method: 'POST',
      data: {
        targets: [request]
      },
    }));

    const nameIdx: number = 0
    const typeIdx: number = 1

    var results: MetricFindValue[] = []
    response.data.results.search.frames.forEach((data: {data: any, schema: any}) => {
      new DataFrameView(toDataFrame(data)).forEach((row) => {
          results.push({text: row[nameIdx], value: row[typeIdx]})
      })
    });

    return new Promise<MetricFindValue[]>((resolve) => {
      resolve(results)
    });
  }

  buildQueryParameters(options: DataQueryRequest<CassandraQuery>): DataQueryRequest<CassandraQuery> {
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
      };
    });

    options.targets = targets;
    
    return options;
  }
}
