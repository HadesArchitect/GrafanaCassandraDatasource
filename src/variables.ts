import { CustomVariableSupport, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { from, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import type { CassandraDatasource } from './datasource';
import { CassandraVariableQuery, normalizeVariableQuery } from './models';
import { VariableQueryEditor } from './VariableQueryEditor';

/**
 * Query variable support for the data source, replacing the deprecated
 * `DataSourcePlugin.setVariableQueryEditor`.
 *
 * The variable query is a CQL statement run through the backend `variables`
 * resource: the first returned column is the value and the optional second
 * column its label.
 */
export class CassandraVariableSupport extends CustomVariableSupport<CassandraDatasource, CassandraVariableQuery> {
  editor = VariableQueryEditor;

  constructor(private readonly datasource: CassandraDatasource) {
    super();
  }

  // An arrow function on purpose: Grafana may pass `variables.query` around
  // detached from this object, so it must not depend on how it is called.
  query = (request: DataQueryRequest<CassandraVariableQuery>): Observable<DataQueryResponse> => {
    // The target is whatever the dashboard stored, which may predate this API.
    const target = normalizeVariableQuery(request.targets[0]);

    return from(this.datasource.metricFindQuery(target, { scopedVars: request.scopedVars, range: request.range })).pipe(
      map((data) => ({ data }))
    );
  };
}
