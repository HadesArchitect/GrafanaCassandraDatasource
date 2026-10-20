---
'grafana-cassandra-datasource': major
---

**Grafana 10.4 is now the minimum supported version.** `grafanaDependency` moves to `>=10.4.0`. The plugin is built against the Grafana 10.4 packages, and the floor now says so. Declaring it lets the plugin drop the deprecated APIs it still carried for older Grafana versions (the legacy variable query editor registration and the legacy secret form field) in favour of the ones Grafana maintains today, which keeps the plugin working as Grafana removes the old ones.

Instances on Grafana 7.4 - 10.3 are unaffected rather than broken: nothing is withdrawn, existing installations keep running the version they already have, and 3.3.0 stays published for them. What they lose is future releases, since 4.0.0 onwards declares Grafana 10.4 as its floor.

Queries, variables, annotations, alerting and every provisioning key behave exactly as before, so upgrading needs no migration. Dashboard variables saved with earlier versions of the plugin keep working as they are, without re-saving the dashboard.

Visible changes in the UI:

- **Stored secrets are shown as "configured".** The password and the TLS certificate content fields in the data source settings no longer appear empty once saved. They show that a value is configured, and a **Reset** button clears it so a new one can be entered. Leaving them alone keeps the stored value, as before.
- The password field now has its own **Password** label, and the variable query editor uses the standard Grafana field layout with a **CQL query** label.
