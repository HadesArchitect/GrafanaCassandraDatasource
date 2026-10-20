import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
} from '@grafana/data';
import {
  CassandraQuery,
  CassandraVariableQuery,
  CassandraDataSourceOptions,
  isQueryComplete,
  normalizeVariableQuery,
} from './models';
import { CassandraVariableSupport } from './variables';
import { Observable } from 'rxjs';

// Default target text of queries created by the old Angular query editor.
const PLACEHOLDER_TARGET = 'select metric';

export class CassandraDatasource extends DataSourceWithBackend<CassandraQuery, CassandraDataSourceOptions> {
  private keyspaces: string[] = [];
  private tables: Map<string, string[]> = new Map();
  private columns: Map<string, string[]> = new Map();

  constructor(instanceSettings: DataSourceInstanceSettings<CassandraDataSourceOptions>) {
    super(instanceSettings);

    // Annotations use the regular query editor and query path.
    // https://grafana.com/developers/plugin-tools/how-to-guides/data-source-plugins/add-support-for-annotation-queries
    this.annotations = {};

    this.variables = new CassandraVariableSupport(this);
  }

  query(options: DataQueryRequest<CassandraQuery>): Observable<DataQueryResponse> {
    // DataSourceWithBackend.query drops every target filterQuery rejects, and answers with
    // no data, rather than an error, when none is left.
    return super.query(this.buildQueryParameters(options));
  }

  /**
   * Decides, for each target on its own, whether it is sent to the backend. Hidden targets,
   * leftover placeholders and queries that are still being written are skipped silently, so a
   * new panel shows no error and one incomplete query does not block the others in the panel.
   */
  filterQuery(query: CassandraQuery): boolean {
    if (query.hide || query.target === PLACEHOLDER_TARGET) {
      return false;
    }

    return isQueryComplete(query);
  }

  // Runs a variable query through the backend `variables` resource. The backend
  // returns MetricFindValue-compatible objects: the first column of each row is the
  // value and the optional second column its label (text).
  // Grafana reaches this through CassandraVariableSupport.
  async metricFindQuery(
    query: CassandraVariableQuery | string,
    options?: LegacyMetricFindQueryOptions
  ): Promise<MetricFindValue[]> {
    const interpolatedQuery = getTemplateSrv().replace(normalizeVariableQuery(query).query, options?.scopedVars, 'csv');

    try {
      return await this.getResource('variables', { query: interpolatedQuery });
    } catch (error) {
      console.warn('Failed to execute variable query:', interpolatedQuery, error);
      return [];
    }
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

  // Interpolates dashboard variables and keeps only the fields the backend reads. Returns a new
  // request: the one Grafana passes in is left untouched.
  buildQueryParameters(options: DataQueryRequest<CassandraQuery>): DataQueryRequest<CassandraQuery> {
    const targets: CassandraQuery[] = options.targets.map((target) => {
      return {
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
        valueId: getTemplateSrv().replace(target.valueId, options.scopedVars, 'csv'),
        alias: target.alias,
        instant: target.instant,
      };
    });

    return { ...options, targets };
  }
}
