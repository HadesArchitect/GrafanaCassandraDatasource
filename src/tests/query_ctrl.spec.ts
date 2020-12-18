//import { CassandraDatasource } from '../datasource'
//import { QueryCtrl } from 'grafana/app/plugins/sdk'
import { CassandraQueryCtrl } from '../query_ctrl'

describe("Class CassandraQueryCtrl:", () => {
    let cassandraQueryCtrl: CassandraQueryCtrl;

    it("unitTestMethod()", () => {
        const a1 = 1;
        const a2 = 3;
        expect(cassandraQueryCtrl.exampleMethod(a1, a2))
            .toBe(4);
    });

    /*
    let cassandraQueryCtrl: CassandraQueryCtrl;   

    it("getOptions()", () => {
        const keyspace = null;
        const table = null;
        const type = '123';
        expect(cassandraQueryCtrl.getOptions(keyspace, table, type))
            .toBe(Promise.resolve([]));
    });  
    */  

});


describe('Sample test', function() {
    it('Condition is true', function() {
      expect('GSDS').toBe('GSDS');
    });
});

