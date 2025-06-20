import _ from 'lodash';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {DataQueryRequest, DataQueryResponse, DataSourceInstanceSettings} from '@grafana/data';
import { CassandraQuery,CassandraVariableQuery, CassandraDataSourceOptions } from './models';
import { Observable } from 'rxjs';

export class CassandraDatasource extends DataSourceWithBackend<CassandraQuery, CassandraDataSourceOptions> {
  headers: any;
  id: number;
  private keyspaces: string[] = [];
  private tables: Map<string, string[]> = new Map();
  private columns: Map<string, string[]> = new Map();

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

  // metricFindQuery implicitly returns array of MetricFindValue objects. It assumed, that
  // backend returns some compatible type to be correctly used in dashboard variables context.
  // https://github.com/grafana/grafana/blob/main/packages/grafana-data/src/types/datasource.ts#L595
  async metricFindQuery(query: CassandraVariableQuery, options?: any) {
    const response = await this.getResource('variables', {query: query.query});
    return response;
  }

  isEditorMode(options: DataQueryRequest<CassandraQuery>): boolean {
    return !options.targets[0].rawQuery;
  }

  isEditorCompleted(options: DataQueryRequest<CassandraQuery>): boolean {
    return Boolean(
      options.targets[0].keyspace &&
      options.targets[0].table &&
      options.targets[0].columnTime &&
      options.targets[0].columnValue &&
      options.targets[0].columnId &&
      options.targets[0].valueId
    );
  }

  isConfiguratorCompleted(options: DataQueryRequest<CassandraQuery>): boolean {
    return Boolean(options.targets[0].target);
  }

  async getKeyspaces(): Promise<string[]> {
    if (0 !== this.keyspaces.length) {
      return this.keyspaces;
    }

    try {
      this.keyspaces = await this.getResource('keyspaces');
      return this.keyspaces;
    } catch (error) {
      console.warn('Failed to fetch keyspaces:', error);
      return [];
    }
  }

  async getTables(keyspace: string): Promise<string[]> {
    if (this.tables.has(keyspace)) {
      return this.tables.get(keyspace)!;
    }

    try {
      const tables = await this.getResource('tables', { keyspace: keyspace });
      this.tables.set(keyspace, tables);
      return tables;
    } catch (error) {
      console.warn(`Failed to fetch tables for keyspace '${keyspace}':`, error);
      return [];
    }
  }

  async getColumns(keyspace: string, table: string, needType: string): Promise<string[]> {
    const cacheKey = `${keyspace}.${table}.${needType}`;
    
    if (this.columns.has(cacheKey)) {
      return this.columns.get(cacheKey)!;
    }

    try {
      const columns = await this.getResource('columns', {
        keyspace: keyspace,
        table: table,
        needType: needType,
      });
      this.columns.set(cacheKey, columns);
      return columns;
    } catch (error) {
      console.warn(`Failed to fetch columns for keyspace '${keyspace}', table '${table}', type '${needType}':`, error);
      return [];
    }
  }

  buildQueryParameters(options: DataQueryRequest<CassandraQuery>): DataQueryRequest<CassandraQuery> {
    //remove placeholder targets
    options.targets = _.filter(options.targets, (target) => {
      return target.target !== 'select metric';
    });

    const targets: CassandraQuery[] = _.map(options.targets, (target) => {
      return {
        datasourceId: target.datasourceId,
        queryType: target.queryType,

        target: getTemplateSrv().replace(target.target, options.scopedVars, 'csv'),
        refId: target.refId,
        hide: target.hide,
        rawQuery: target.rawQuery,
        filtering: target.filtering,
        keyspace: target.keyspace,
        table: target.table,
        columnTime: target.columnTime,
        columnValue: target.columnValue,
        columnId: target.columnId,
        valueId:  getTemplateSrv().replace(target.valueId, options.scopedVars, 'csv'),
        alias: target.alias,
        instant: target.instant,
      };
    });

    options.targets = targets;

    return options;
  }
}
