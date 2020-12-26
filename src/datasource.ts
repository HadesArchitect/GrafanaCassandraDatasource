import _ from "lodash";
import {TSDBRequest, TSDBQuery, TSDBRequestOptions, TableMetadata} from './models';

export class CassandraDatasource {
  name: string;
  type: string;
  id: string;
  url: string;
  keyspace: string;
  user: string;
  withCredentials: boolean;
  instanceSettings: any;
  headers: any;

  /** @ngInject */
  constructor(instanceSettings, private backendSrv, private templateSrv) {
    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.keyspace = instanceSettings.jsonData.keyspace;
    this.user = instanceSettings.jsonData.user;
    this.name = instanceSettings.name;
    this.id = instanceSettings.id;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = {'Content-Type': 'application/json'};
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }
  }

  query(options: any) {
    const query = this.buildQueryParameters(options);
    query.targets = query.targets.filter(t => !t.hide);

    if (query.targets.length <= 0) {
      return Promise.resolve({data: []});
    }

    return this.doTsdbRequest(query).then(handleTsdbResponse);
  }

  testDatasource(): {} {
    return this.backendSrv
      .datasourceRequest({
        url: '/api/tsdb/query',
        method: 'POST',
        data: {
          from: '5m',
          to: 'now',
          queries: [{ datasourceId: this.id, queryType: 'connection', keyspace: this.keyspace }]
        },
      })
      .then(() => {
        return { status: 'success', message: 'Database Connection OK' };
      })
      .catch((error: any) => {
        return { status: 'error', message: error.data.message };
      });
  }

  metricFindQuery(keyspace: string, table: string): TableMetadata {
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
  }

  doRequest(options) {
    options.withCredentials = this.withCredentials;
    options.headers = this.headers;

    return this.backendSrv.datasourceRequest(options);
  }

  doTsdbRequest(options: TSDBRequestOptions) {
    const tsdbRequestData: TSDBRequest = {
      queries: options.targets,
    };

    if (options.range) {
      tsdbRequestData.from = options.range.from.valueOf().toString();
      tsdbRequestData.to = options.range.to.valueOf().toString();
    }

    return this.backendSrv.datasourceRequest({
      url: '/api/tsdb/query',
      method: 'POST',
      data: tsdbRequestData
    });
  }

  buildQueryParameters(options: any): TSDBRequestOptions {
    //remove placeholder targets
    options.targets = _.filter(options.targets, target => {
      return target.target !== 'select metric';
    });

    const targets = _.map(options.targets, target => {
      return {
        queryType: 'query',
        target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        rawQuery: target.rawQuery,
        type: target.type || 'timeserie',
        datasourceId: this.id,
        filtering: target.filtering,
        keyspace: target.keyspace,
        table: target.table,
        columnTime: target.columnTime,
        columnValue: target.columnValue,
        columnId: target.columnId,
        valueId: target.valueId
      };
    });

    options.targets = targets;

    return options;
  }
}

export function handleTsdbResponse(response) {
  const res : object[] = [];
  _.forEach(response.data.results, r => {
    _.forEach(r.series, s => {
      res.push({target: s.name, datapoints: s.points});
    });
    _.forEach(r.tables, t => {
      t.type = 'table';
      t.refId = r.refId;
      res.push(t);
    });
  });
  response.data = res;

  return response;
}

export function mapToTextValue(result) {
  return _.map(result, (d, i) => {
    if (d && d.text && d.value) {
      return { text: d.text, value: d.value };
    } else if (_.isObject(d)) {
      return { text: d, value: i};
    }
    return { text: d, value: d };
  });
}
