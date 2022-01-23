import _ from 'lodash';
import {
  DataSourceWithBackend,
  getBackendSrv,
  getTemplateSrv,
  FetchResponse,
  FetchError,
  FetchErrorDataProps,
} from '@grafana/runtime';
import { DataQueryRequest, DataQueryResponse, DataSourcePluginMeta, DataSourceJsonData, DataSourceInstanceSettings } from '@grafana/data';
import { TSDBRequest, CassandraQuery, TSDBRequestOptions /*TableMetadata*/ } from './models';

import { Observable, of } from 'rxjs';
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
   /*  const query = this.buildQueryParameters(options);
    query.targets = query.targets.filter((t) => !t.hide); */

    console.log(options);
    if (options.targets.length <= 0) {
      return of({ data: [] });
    }

    return this.doTsdbRequest(options)
  }

  async testDatasource(): Promise<any> {
    return getBackendSrv()
      .fetch({
        url: '/api/tsdb/query',
        method: 'POST',
        data: {
          from: '5m',
          to: 'now',
          queries: [{ datasourceId: this.id, queryType: 'connection', keyspace: this.keyspace }],
        },
      })
      .subscribe(
        (res: FetchResponse<any>) => {
          console.log(res);
          return { status: 'success', message: 'Database connected', details: {} };
        },
        (err: FetchError<FetchErrorDataProps>) => {
          console.log(err);
          return { status: 'error', message: err.statusText, details: {} };
        }
      );
  }

  /*metricFindQuery(keyspace: string, table: string): TableMetadata {
    const interpolated: TSDBQuery = {
      datasourceId: this.id,
      queryType: "search",
      refId: "search",
      keyspace: keyspace,
      table: table
    };
    
    return this.doTsdbRequest({
      targets: [interpolated]
    }).then(response => {
      const tmd = new TableMetadata(response.data.results.search.tables["0"].rows["0"]["0"]);
      // return tmd.toSuggestion();
      return tmd;
    }).catch((error: any) => {
      console.log(error);
      return new TableMetadata();
    });
  }*/

  doRequest(options) {
    options.headers = this.headers;

    return getBackendSrv().fetch(options);
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
      url: '/api/tsdb/query',
      method: 'POST',
      data: tsdbRequestData,
    });
  }

  buildQueryParameters(options: any): TSDBRequestOptions {
    //remove placeholder targets
    options.targets = _.filter(options.targets, (target) => {
      return target.target !== 'select metric';
    });

    //console.log(options.targets);
    const targets = _.map(options.targets, (target) => {
      return {
        queryType: 'query',
        target: getTemplateSrv().replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        rawQuery: target.rawQuery,
        //type: target.type || 'timeserie',
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
    console.log(options.targets);
    return options;
  }
}

export function handleTsdbResponse(response) {
  console.log(response);
  console.log('data');
  console.log(response.data);
  console.log('responses');
  console.log(response.data.responses);
  console.log('results');
  console.log(response.data.results);

  const res: object[] = [];
  //_.forEach(response.data.results, r => {

  for (const key in response.data.results) {
    response.data.results[key].refId = key;
    //res.push(response.data.results[key]);
    console.log(response.data.results[key]);

    response.data.results[key].dataframes.forEach((value) => {
      console.log(value);
      //value.refId = key;
      res.push(value);
    });
  }

  /*var frames = new Map<string, DataFrame>(JSON.parse(response.data));
    console.log(frames);
    frames.forEach((value: DataFrame, key: string, frames) => {
      console.log("value");
      console.log(value);
      console.log("key");
      console.log(key);
      value.refId = key;
      res.push(value);
    });*/

  /*_.forEach(r.series, s => {
      res.push({target: s.name, datapoints: s.points});
    });
    _.forEach(r.tables, t => {
      t.type = 'table';
      t.refId = r.refId;
      res.push(t);
    });*/
  //});
  response.data = res;

  console.log(response);

  return response;
}

export function mapToTextValue(result) {
  return _.map(result, (d, i) => {
    if (d && d.text && d.value) {
      return { text: d.text, value: d.value };
    } else if (_.isObject(d)) {
      return { text: d, value: i };
    }
    return { text: d, value: d };
  });
}
