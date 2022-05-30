# Conduit Connector Zendesk

### General

The Zendesk connector is one of [Conduit](https://github.com/ConduitIO/conduit) custom plugins. It provides both, a source
and a destination zendesk connectors.

### How to build it

Run `make`.

### Testing

Run `make test` to run all the tests.

## Zendesk Source

The Zendesk client connector will connect with Zendesk API through the `url` constructed using subdomain specific to individual organization. Upon successful configuration with `zendesk.userName ` and `zendesk.apiToken` the tickets from the given domain is fetched using cursor based [incremental exports](https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/) provided by zendesk. The cursor is initiated with start_time set to `0` or the time set in `position` of last successfully ack'd record and all subsequent iterations are done using `after_url` returned by the zendesk till the pipeline is paused. On resuming of the pipeline updated_at time of the last fetched ticket is used to restart the cursor.

### Generating API token in Zendesk
The api token for the zendesk can be created through zendesk portal by logging in as admin. Refer the zendesk [documentation](https://support.zendesk.com/hc/en-us/articles/4408889192858-Generating-a-new-API-token#topic_bsw_lfg_mmb) for step-by-step setup.

### Change Data Capture (CDC)
The connector uses the zendesk [cursor based incremental exports](https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/) to listen iterate over tickets changed after the given `start_time`. 
We initiate a `cursor` at the start of the pipeline using the `start_time` as 0, which means we start fetching all the tickets from the start. The subsequent data is fetched using the `after_url` received as part of response.
When the pipeline resumed after pause/crash, we use the position of the last successfully read record to restart the cursor using the updated_at data from position as the start_time.  


#### Position Handling

The connector uses the combination of `last_modified_time` time and `id` to uniquely identify the records.
`last_modified_time`: The `updated_at` time of last successfully read ticket is used. In case the `updated_at` time is empty, 
the `created_at` time of the last ticket is used.
`id`: This is the ticket id associated with the ticket, received from zendesk.

The `last_modified_time` is used as `start_time` query param for restarting the cursor based incremental export.  

Sample position:
```json
{
  "last_modified_time": "2006-01-02T15:04:05Z07:00",
  "id": 12345
}
```

### Record Keys

The `id` of the ticket is used as the unique key for the record.

Sample Record:
```json
{
  "position": {
    "last_modified_time": "2006-01-02T15:04:05Z07:00",
    "id": 12345
  },
  "metadata": null,
  "created_at": "2006-01-02T15:04:05Z07:00",
  "key": "12345",
  "payload": "<ticket json received from zendesk>"
}
```

### Configuration - Source

| name                  | description                                                                  | required | default |
| -------               |------------------------------------------------------------------------------| -------- |---------|
|`zendesk.domain`       | domain is the registered by organization to zendesk                          | true     |         |
|`zendesk.userName`     | username is the registered for login                                         | true     |         |
|`zendesk.apiToken`     | password associated with the username for login                              | true     |         |
|`pollingPeriod`        | pollingPeriod is the frequency of conduit hitting zendesk API- Default is 6s | false    | "6s"    |

**NOTE:** `pollingPeriod` will be in time.Duration - `2ns`,`2ms`,`2s`,`2m`,`2h`

### Known Limitations

* The zendesk API has a rate limit of 10 requests per minute. If rate limit is exceeded, zendesk sends 429 status code with Cool off duration in `Retry-After` header.
We use this duration to skip hitting the zendesk APIs repeatedly.
* Currently, the connector only supports ticket data fetching. Other type of data fetching will be part of subsequent phases.
