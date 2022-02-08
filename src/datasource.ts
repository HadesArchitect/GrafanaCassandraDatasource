import _ from 'lodash';
import {
  DataSourceWithBackend,
  getBackendSrv,
  getTemplateSrv,
  FetchResponse,
} from '@grafana/runtime';
import {
  toDataFrame,
  DataFrameView,
  DataQueryRequest,
  DataQueryResponse,
  DataSourcePluginMeta,
  DataSourceJsonData,
  DataSourceInstanceSettings,
  MetricFindValue,
} from '@grafana/data';
import { TSDBRequest, CassandraQuery, TSDBRequestOptions /*TableMetadata*/ } from './models';
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
  keyspace: string;
  headers: any;
  name: string;
  id: number;
  meta: DataSourcePluginMeta;

  /** @ngInject */
  constructor(instanceSettings: DataSourceInstanceSettings<CassandraDataSourceOptions>) {
    super(instanceSettings);
    this.keyspace = instanceSettings.jsonData.keyspace;
    this.headers = { 'Content-Type': 'application/json' };
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }

    this.name = instanceSettings.name;
    this.id = instanceSettings.id;
    this.meta = instanceSettings.meta;
  }

  query(options: DataQueryRequest<CassandraQuery>): Observable<DataQueryResponse> {
    const query = this.buildQueryParameters(options);
    query.targets = query.targets.filter((t) => !t.hide);

    return super.query(options)
  }

/*   async testDatasource(): Promise<any> {
    this.metricFindQuery("test", "test")

    return getBackendSrv()
      .fetch({
        url: '/api/ds/query',
        method: 'POST',
        data: {
          from: '5m',
          to: 'now',
          queries: [{ datasourceId: this.id, queryType: 'connection', keyspace: this.keyspace }],
        },
      })
      .toPromise()
      .then(() => {
        return { status: 'success', message: 'Database Connection OK' };
      })
      .catch((error: FetchError) => {
        return { status: 'error', message: exctractErrors(error) };
      });
  } */

  
  async metricFindQuery(keyspace: string, table: string): Promise<MetricFindValue[]> {  
    const request: CassandraQuery = {
      datasourceId: this.id,
      queryType: "search",
      refId: "search",
      keyspace: keyspace,
      table: table
    };
    
   
    var response = await lastValueFrom(this.doTsdbRequest({
      targets: [request]
    }))

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

    /* this.doTsdbRequest({
      targets: [request]
    })
      .toPromise()
      .then(response => {
        const nameIdx: number = 0
        const typeIdx: number = 1

        response.data.results.search.frames.forEach((data: {data: any, schema: any}) => {
          new DataFrameView(toDataFrame(data)).forEach((row) => {
              results.push({text: row[nameIdx], value: row[typeIdx]})
          })
        });
      }).catch((error: any) => {
        console.log(error);
      });


     
      console.log(results)
      return new Promise<MetricFindValue[]>((resolve) => {
          console.log({results: results});
          resolve(results)
      }); */

  }


  doTsdbRequest(options: TSDBRequestOptions): Observable<FetchResponse<any>> {
    const tsdbRequestData: TSDBRequest = {
      queries: options.targets,
    };

    if (options.range) {
      tsdbRequestData.from = options.range.from.valueOf().toString();
      tsdbRequestData.to = options.range.to.valueOf().toString();
    }

    return getBackendSrv().fetch({
      url: '/api/ds/query',
      method: 'POST',
      data: tsdbRequestData,
    });
  }

  buildQueryParameters(options: any): TSDBRequestOptions {
    //remove placeholder targets
    options.targets = _.filter(options.targets, (target) => {
      return target.target !== 'select metric';
    });

    const targets = _.map(options.targets, (target) => {
      return {
        datasourceId: target.datasourceId,
        queryType: 'query',
        
        target: getTemplateSrv().replace(target.target, options.scopedVars, 'regex'),
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
    //console.log(options.targets);
    return options;
  }
}

/* function exctractErrors(response: FetchError): string {
  var result: string = ""
  for (let key of Object.keys(response.data.results)) {
    result = "Query " + key + ": " + response.data.results[key].error + ", "
  }
  return result.substring(0, result.length - 2);
} */
