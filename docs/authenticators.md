# Custom Authenticators (LDAP, etc.)

Some Cassandra clusters authenticate clients with a non-standard authenticator —
for example `org.apache.cassandra.auth.LDAPAuthenticator`. By default the driver
used by this plugin (gocql) only accepts a fixed allow-list of server-side
authenticator class names, and refuses to connect to anything else with an error
like:

```text
Failed to create Cassandra connection
cluster.CreateSession: gocql: unable to create session: unable to discover protocol version:
unexpected authenticator "org.apache.cassandra.auth.LDAPAuthenticator"
```

This happens during protocol negotiation, *before* your credentials are even
sent — so it is not a username/password problem.

## The `Allowed authenticators` setting

The **Allowed authenticators** connection setting lets you tell the driver which
server-side authenticators it is allowed to authenticate against. Add your
cluster's authenticator class name and the connection will succeed.

* It is a **semicolon-separated** list of fully-qualified authenticator class names.
* **Leaving it empty preserves the original behaviour** — the driver falls back
  to its built-in default allow-list (see below). Existing data sources are
  therefore unaffected.
* The list **replaces** the default list rather than extending it. If your
  cluster can negotiate more than one authenticator, include each one you want to
  allow.


## Configuring in the UI

Open the data source settings and, under **Connection settings**, fill in
**Allowed authenticators**, e.g.:

```
org.apache.cassandra.auth.PasswordAuthenticator;org.apache.cassandra.auth.LDAPAuthenticator
```

Then **Save & test**.

## Configuring via provisioning

Set the `allowedAuthenticators` key under `jsonData`:

```yaml
apiVersion: 1
datasources:
  - name: "Cassandra (LDAP)"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "cassandra:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
      allowedAuthenticators: "org.apache.cassandra.auth.PasswordAuthenticator;org.apache.cassandra.auth.LDAPAuthenticator"
    secureJsonData:
      password: cassandra
```

## Default allow-list

When **Allowed authenticators** is left empty, the driver accepts the following
authenticators (its built-in defaults):

```
org.apache.cassandra.auth.PasswordAuthenticator
com.instaclustr.cassandra.auth.SharedSecretAuthenticator
com.datastax.bdp.cassandra.auth.DseAuthenticator
io.aiven.cassandra.auth.AivenAuthenticator
com.ericsson.bss.cassandra.ecaudit.auth.AuditPasswordAuthenticator
com.amazon.helenus.auth.HelenusAuthenticator
com.ericsson.bss.cassandra.ecaudit.auth.AuditAuthenticator
com.scylladb.auth.SaslauthdAuthenticator
com.scylladb.auth.TransitionalAuthenticator
com.instaclustr.cassandra.auth.InstaclustrPasswordAuthenticator
```

If your cluster uses one of these, you do not need to set anything. If it uses
something else (such as `LDAPAuthenticator`), add it — and remember the list
replaces the defaults, so include `PasswordAuthenticator` too if you still need
it.
